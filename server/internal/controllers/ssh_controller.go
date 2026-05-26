package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cmdb-server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SSHController struct {
	sshService *services.SSHService
}

type sshTestRequest struct {
	AssetID uint   `json:"assetId"`
	Command string `json:"command"`
}

type terminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewSSHController(sshService *services.SSHService) *SSHController {
	return &SSHController{sshService: sshService}
}

func (ctl *SSHController) Test(c *gin.Context) {
	var req sshTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if req.AssetID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assetId 不能为空"})
		return
	}

	output, asset, err := ctl.sshService.TestCommand(req.AssetID, req.Command)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	command := strings.TrimSpace(req.Command)
	if command == "" {
		command = "hostname && uptime"
	}

	c.JSON(http.StatusOK, gin.H{
		"assetId":    asset.ID,
		"hostname":   asset.Hostname,
		"address":    asset.Address,
		"command":    command,
		"output":     output,
		"executedAt": time.Now().Format(time.RFC3339),
	})
}

func (ctl *SSHController) TerminalWS(c *gin.Context) {
	assetIDRaw := strings.TrimSpace(c.Query("assetId"))
	assetIDNum, err := strconv.ParseUint(assetIDRaw, 10, 64)
	if err != nil || assetIDNum == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assetId 参数无效"})
		return
	}

	conn, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client, asset, err := ctl.sshService.DialClient(uint(assetIDNum))
	if err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: err.Error()})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("创建 SSH 会话失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("打开 stdin 失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("打开 stdout 失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("打开 stderr 失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 40, 120, modes); err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("申请 PTY 失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}
	if err := session.Shell(); err != nil {
		_ = conn.WriteJSON(terminalMessage{Type: "error", Data: fmt.Sprintf("启动 shell 失败: %v", err)})
		_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
		return
	}

	_ = conn.WriteJSON(terminalMessage{
		Type: "ready",
		Data: fmt.Sprintf("已连接到 %s (%s:%d)", asset.Hostname, asset.Address, asset.Port),
	})

	writeCh := make(chan terminalMessage, 128)
	done := make(chan struct{})
	var closeOnce sync.Once

	stop := func() {
		closeOnce.Do(func() {
			close(done)
			close(writeCh)
			_ = stdin.Close()
			_ = session.Close()
		})
	}

	go func() {
		for msg := range writeCh {
			if err := conn.WriteJSON(msg); err != nil {
				stop()
				return
			}
		}
	}()

	forward := func(reader io.Reader) {
		buf := make([]byte, 2048)
		for {
			n, readErr := reader.Read(buf)
			if n > 0 {
				select {
				case writeCh <- terminalMessage{Type: "output", Data: string(buf[:n])}:
				case <-done:
					return
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					select {
					case writeCh <- terminalMessage{Type: "error", Data: readErr.Error()}:
					case <-done:
					}
				}
				return
			}
		}
	}

	go forward(stdout)
	go forward(stderr)

	for {
		_, raw, readErr := conn.ReadMessage()
		if readErr != nil {
			break
		}

		var msg terminalMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			text := string(raw)
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			if _, err := io.WriteString(stdin, text); err != nil {
				writeCh <- terminalMessage{Type: "error", Data: fmt.Sprintf("写入远端失败: %v", err)}
				break
			}
			continue
		}

		switch msg.Type {
		case "input":
			text := msg.Data
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			if _, err := io.WriteString(stdin, text); err != nil {
				writeCh <- terminalMessage{Type: "error", Data: fmt.Sprintf("写入远端失败: %v", err)}
				break
			}
		case "resize":
			rows := msg.Rows
			cols := msg.Cols
			if rows <= 0 {
				rows = 40
			}
			if cols <= 0 {
				cols = 120
			}
			if err := session.WindowChange(rows, cols); err != nil {
				writeCh <- terminalMessage{Type: "error", Data: fmt.Sprintf("窗口调整失败: %v", err)}
			}
		case "signal":
			if msg.Data == "ctrl_c" {
				if _, err := io.WriteString(stdin, "\u0003"); err != nil {
					writeCh <- terminalMessage{Type: "error", Data: fmt.Sprintf("发送 Ctrl+C 失败: %v", err)}
				}
			}
		}
	}

	stop()
	_ = conn.WriteJSON(terminalMessage{Type: "closed", Data: "会话已关闭"})
}
