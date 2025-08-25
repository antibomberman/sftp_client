package sftp_client

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	client *sftp.Client
	conn   *ssh.Client
}

func Connect(host, user, password string) (*SFTPClient, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Используйте более безопасный метод в продакшене
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %v", err)
	}

	// Используем NewClient вместо sftp.New
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("не удалось создать SFTP клиент: %v", err)
	}

	return &SFTPClient{
		client: client,
		conn:   conn,
	}, nil
}

func ConnectWithKey(host, user, keyPath string) (*SFTPClient, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать приватный ключ: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить приватный ключ: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %v", err)
	}

	// Используем NewClient вместо sftp.New
	client, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("не удалось создать SFTP клиент: %v", err)
	}

	return &SFTPClient{
		client: client,
		conn:   conn,
	}, nil
}

func (s *SFTPClient) Close() error {
	if s.client != nil {
		s.client.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

// CREATE OPERATIONS

// UploadFile загружает локальный файл на удаленный сервер
func (s *SFTPClient) UploadFile(localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть локальный файл: %v", err)
	}
	defer localFile.Close()

	// Создаем директорию на сервере, если она не существует
	dir := filepath.Dir(remotePath)
	if err := s.client.MkdirAll(dir); err != nil {
		return fmt.Errorf("не удалось создать директорию на сервере: %v", err)
	}

	remoteFile, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("не удалось создать удаленный файл: %v", err)
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("не удалось загрузить файл: %v", err)
	}

	return nil
}

// CreateDirectory создает директорию на сервере
func (s *SFTPClient) CreateDirectory(remotePath string) error {
	return s.client.MkdirAll(remotePath)
}

// CreateFile создает пустой файл на сервере
func (s *SFTPClient) CreateFile(remotePath string) error {
	file, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %v", err)
	}
	return file.Close()
}

// READ OPERATIONS

// DownloadFile скачивает файл с сервера
func (s *SFTPClient) DownloadFile(remotePath, localPath string) error {
	remoteFile, err := s.client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть удаленный файл: %v", err)
	}
	defer remoteFile.Close()

	// Создаем локальную директорию, если она не существует
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("не удалось создать локальную директорию: %v", err)
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("не удалось создать локальный файл: %v", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("не удалось скачать файл: %v", err)
	}

	return nil
}

// ReadFileContent читает содержимое файла в память
func (s *SFTPClient) ReadFileContent(remotePath string) ([]byte, error) {
	file, err := s.client.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// ListDirectory выводит список файлов в директории
func (s *SFTPClient) ListDirectory(remotePath string) ([]os.FileInfo, error) {
	files, err := s.client.ReadDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию: %v", err)
	}
	return files, nil
}

// GetFileInfo получает информацию о файле
func (s *SFTPClient) GetFileInfo(remotePath string) (os.FileInfo, error) {
	info, err := s.client.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить информацию о файле: %v", err)
	}
	return info, nil
}

// FileExists проверяет существование файла
func (s *SFTPClient) FileExists(remotePath string) (bool, error) {
	_, err := s.client.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UPDATE OPERATIONS

// UpdateFile обновляет содержимое файла (перезаписывает)
func (s *SFTPClient) UpdateFile(remotePath string, content []byte) error {
	file, err := s.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл для записи: %v", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("не удалось записать в файл: %v", err)
	}

	return nil
}

// AppendToFile добавляет содержимое в конец файла
func (s *SFTPClient) AppendToFile(remotePath string, content []byte) error {
	file, err := s.client.OpenFile(remotePath, os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл для добавления: %v", err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("не удалось добавить в файл: %v", err)
	}

	return nil
}

// RenameFile переименовывает файл
func (s *SFTPClient) RenameFile(oldPath, newPath string) error {
	return s.client.Rename(oldPath, newPath)
}

// DELETE OPERATIONS

// DeleteFile удаляет файл
func (s *SFTPClient) DeleteFile(remotePath string) error {
	return s.client.Remove(remotePath)
}

// DeleteDirectory удаляет директорию
func (s *SFTPClient) DeleteDirectory(remotePath string) error {
	return s.client.RemoveDirectory(remotePath)
}

// DeleteDirectoryRecursive рекурсивно удаляет директорию со всем содержимым
func (s *SFTPClient) DeleteDirectoryRecursive(remotePath string) error {
	return s.client.RemoveAll(remotePath)
}
