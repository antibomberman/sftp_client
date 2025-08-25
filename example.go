package sftp_client

import (
	"fmt"
	"log"
)

// Пример использования
func Example() {
	// Подключение к SFTP серверу
	client, err := Connect("your-sftp-server.com:22", "username", "password")
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer client.Close()

	// CREATE
	fmt.Println("=== CREATE OPERATIONS ===")
	err = client.CreateDirectory("/test/newdir")
	if err != nil {
		log.Println("Ошибка создания директории:", err)
	} else {
		fmt.Println("Директория создана")
	}

	err = client.CreateFile("/test/newfile.txt")
	if err != nil {
		log.Println("Ошибка создания файла:", err)
	} else {
		fmt.Println("Файл создан")
	}

	// UPLOAD
	err = client.UploadFile("./local_file.txt", "/test/uploaded_file.txt")
	if err != nil {
		log.Println("Ошибка загрузки файла:", err)
	} else {
		fmt.Println("Файл загружен")
	}

	// READ
	fmt.Println("\n=== READ OPERATIONS ===")
	content := []byte("Hello, SFTP!")
	err = client.UpdateFile("/test/newfile.txt", content)
	if err != nil {
		log.Println("Ошибка записи в файл:", err)
	}

	readContent, err := client.ReadFileContent("/test/newfile.txt")
	if err != nil {
		log.Println("Ошибка чтения файла:", err)
	} else {
		fmt.Printf("Содержимое файла: %s\n", string(readContent))
	}

	files, err := client.ListDirectory("/test")
	if err != nil {
		log.Println("Ошибка получения списка файлов:", err)
	} else {
		fmt.Println("Файлы в директории:")
		for _, file := range files {
			fmt.Printf("  %s (%d bytes)\n", file.Name(), file.Size())
		}
	}

	// UPDATE
	fmt.Println("\n=== UPDATE OPERATIONS ===")
	err = client.AppendToFile("/test/newfile.txt", []byte("\nAppended text"))
	if err != nil {
		log.Println("Ошибка добавления в файл:", err)
	} else {
		fmt.Println("Текст добавлен в файл")
	}

	err = client.RenameFile("/test/newfile.txt", "/test/renamed_file.txt")
	if err != nil {
		log.Println("Ошибка переименования файла:", err)
	} else {
		fmt.Println("Файл переименован")
	}

	// DELETE
	fmt.Println("\n=== DELETE OPERATIONS ===")
	err = client.DeleteFile("/test/renamed_file.txt")
	if err != nil {
		log.Println("Ошибка удаления файла:", err)
	} else {
		fmt.Println("Файл удален")
	}

	err = client.DeleteDirectoryRecursive("/test")
	if err != nil {
		log.Println("Ошибка удаления директории:", err)
	} else {
		fmt.Println("Директория удалена")
	}
}
