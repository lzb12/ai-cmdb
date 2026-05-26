package services

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"cmdb-server/internal/global"
	"cmdb-server/internal/models"
	"golang.org/x/crypto/ssh"
)

type SSHService struct {
	credentialService *CredentialService
}

func NewSSHService(credentialService *CredentialService) *SSHService {
	return &SSHService{credentialService: credentialService}
}

func (s *SSHService) TestCommand(assetID uint, command string) (string, *models.HostAsset, error) {
	asset, credential, err := s.loadAssetAndCredential(assetID)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(command) == "" {
		command = "hostname && uptime"
	}

	output, err := s.execute(*asset, *credential, command)
	if err != nil {
		return "", nil, err
	}
	return output, asset, nil
}

func (s *SSHService) DialClient(assetID uint) (*ssh.Client, *models.HostAsset, error) {
	asset, credential, err := s.loadAssetAndCredential(assetID)
	if err != nil {
		return nil, nil, err
	}

	authMethod, err := s.buildAuthMethod(*credential)
	if err != nil {
		return nil, nil, err
	}

	port := asset.Port
	if port <= 0 {
		port = 22
	}
	config := &ssh.ClientConfig{
		User:            credential.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(asset.Address, strconv.Itoa(port)), config)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH 连接失败: %w", err)
	}
	return client, asset, nil
}

func (s *SSHService) loadAssetAndCredential(assetID uint) (*models.HostAsset, *models.Credential, error) {
	var asset models.HostAsset
	if err := global.DB.First(&asset, assetID).Error; err != nil {
		return nil, nil, err
	}
	if asset.CredentialID == nil || *asset.CredentialID == 0 {
		return nil, nil, fmt.Errorf("资产未绑定凭据")
	}

	credential, err := s.credentialService.FindByID(*asset.CredentialID)
	if err != nil {
		return nil, nil, err
	}
	return &asset, credential, nil
}

func (s *SSHService) execute(asset models.HostAsset, credential models.Credential, command string) (string, error) {
	authMethod, err := s.buildAuthMethod(credential)
	if err != nil {
		return "", err
	}

	port := asset.Port
	if port <= 0 {
		port = 22
	}
	config := &ssh.ClientConfig{
		User:            credential.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         8 * time.Second,
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(asset.Address, strconv.Itoa(port)), config)
	if err != nil {
		return "", fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH 会话失败: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	text := strings.TrimSpace(string(output))
	if len(text) > 4000 {
		text = text[:4000] + "\n...(truncated)"
	}
	if err != nil {
		if text == "" {
			return "", fmt.Errorf("远程命令失败: %w", err)
		}
		return "", fmt.Errorf("远程命令失败: %s", text)
	}
	if text == "" {
		return "(command completed, no output)", nil
	}
	return text, nil
}

func (s *SSHService) buildAuthMethod(credential models.Credential) (ssh.AuthMethod, error) {
	switch credential.AuthType {
	case "password":
		password, err := global.ENCRYPTOR.Decrypt(credential.PasswordEnc)
		if err != nil {
			return nil, fmt.Errorf("密码解密失败")
		}
		if password == "" {
			return nil, fmt.Errorf("密码为空")
		}
		return ssh.Password(password), nil
	case "key":
		privateKey, err := global.ENCRYPTOR.Decrypt(credential.PrivateKeyEnc)
		if err != nil {
			return nil, fmt.Errorf("私钥解密失败")
		}
		if strings.TrimSpace(privateKey) == "" {
			return nil, fmt.Errorf("私钥为空")
		}
		passphrase, err := global.ENCRYPTOR.Decrypt(credential.PassphraseEnc)
		if err != nil {
			return nil, fmt.Errorf("口令解密失败")
		}

		var signer ssh.Signer
		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		return ssh.PublicKeys(signer), nil
	default:
		return nil, fmt.Errorf("未知认证类型")
	}
}
