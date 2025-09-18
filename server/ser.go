package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const serverAddress = "0.0.0.0:50050"

var clients []net.Conn

func startServer(output *TransparentEntry) {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("加载证书失败: %v", err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	listener, err := tls.Listen("tcp", serverAddress, config)
	if err != nil {
		log.Fatalf("TLS 服务器启动失败: %v", err)
	}
	defer listener.Close()

	log.Printf("服务端启动成功，监听端口 %s\n", serverAddress)
	updateOutput(output, "服务端启动成功，等待客户端连接...\n")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("客户端连接失败:", err)
			continue
		}

		clients = append(clients, conn)
		log.Println("新客户端连接:", conn.RemoteAddr())
		updateOutput(output, fmt.Sprintf("客户端 %s 已连接\n", conn.RemoteAddr()))

		go handleClient(conn, output)
	}
}

func handleClient(conn net.Conn, output *TransparentEntry) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("客户端 %s 断开连接: %v\n", clientAddr, err)
			updateOutput(output, fmt.Sprintf("客户端 %s 断开连接\n", clientAddr))
			removeClient(conn)
			return
		}

		message = strings.TrimSpace(message)
		log.Printf("收到 %s 的消息: %s\n", clientAddr, message)
		updateOutput(output, fmt.Sprintf("客户端 %s 返回: %s\n", clientAddr, message))

		if message == "FILE_TRANSFER_TO_SERVER_START" {
			handleFileTransferFromClient(conn, reader, output)
		}

		if message == "FILE_TRANSFER_TO_SER_START" {
			log.Println("开始接收客户端发送的文件...")
			handleFileTransferFromClient(conn, reader, output)
			continue
		}

		if message == "FILE_LIST_START" {
			handleFileListFromClient(reader, output)
		}
	}
}

func handleFileTransferFromClient(conn net.Conn, reader *bufio.Reader, output *TransparentEntry) {
	fileName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading file name: %v\n", err)
		updateOutput(output, fmt.Sprintf("Error reading file name: %v\n", err))
		return
	}
	fileName = strings.TrimSpace(fileName)

	fileSizeStr, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading file size: %v\n", err)
		updateOutput(output, fmt.Sprintf("Error reading file size: %v\n", err))
		return
	}
	fileSizeStr = strings.TrimSpace(fileSizeStr)

	fileSize, err := strconv.Atoi(fileSizeStr)
	if err != nil || fileSize <= 0 {
		log.Printf("Invalid file size: %v\n", err)
		updateOutput(output, fmt.Sprintf("Invalid file size: %v\n", err))
		return
	}

	fileContent := make([]byte, fileSize)
	_, err = io.ReadFull(reader, fileContent)
	if err != nil {
		log.Printf("Error reading file content: %v\n", err)
		updateOutput(output, fmt.Sprintf("Error reading file content: %v\n", err))
		return
	}

	err = ioutil.WriteFile(fileName, fileContent, 0644)
	if err != nil {
		log.Printf("Error saving file: %v\n", err)
		updateOutput(output, fmt.Sprintf("Error saving file: %v\n", err))
	} else {
		log.Printf("文件保存成功: %s\n", fileName)
		updateOutput(output, fmt.Sprintf("文件保存成功: %s\n", fileName))
	}
}

func handleFileListFromClient(reader *bufio.Reader, output *TransparentEntry) {
	var fileList []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("读取文件列表失败: %v\n", err)
			updateOutput(output, fmt.Sprintf("读取文件列表失败: %v\n", err))
			return
		}
		line = strings.TrimSpace(line)
		if line == "FILE_LIST_END" {
			break
		}
		fileList = append(fileList, line)
	}

	// 更新文件浏览页面
	updateFileBrowser(fileList)
}

func updateFileBrowser(fileList []string) {
	// 假设 fileListWidget 是文件浏览页面的列表组件
	fileListWidget := widget.NewList(
		func() int {
			return len(fileList)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(fileList[id])
		},
	)
	fileListWidget.Refresh()
}

func broadcastCommand(command string, output *TransparentEntry) {
	if len(clients) == 0 {
		updateOutput(output, "没有客户端连接，无法发送命令")
		return
	}
	for _, client := range clients {
		_, err := client.Write([]byte(command + "\n"))
		if err != nil {
			log.Println("发送命令失败:", err)
			updateOutput(output, fmt.Sprintf("发送命令失败: %v", err))
		} else {
			updateOutput(output, fmt.Sprintf("命令发送给 %s: %s", client.RemoteAddr(), command))
		}
	}
}

func removeClient(conn net.Conn) {
	for i, client := range clients {
		if client == conn {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func showFileTransferDialog(w fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			updateOutput(resultOutput, "文件选择出错\n")
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		filePath := reader.URI().Path()
		updateOutput(resultOutput, "选择的文件路径: "+filePath+"\n")
		fileContent, err := ioutil.ReadAll(reader)
		if err != nil {
			updateOutput(resultOutput, "读取文件内容失败\n")
			return
		}
		for _, client := range clients {
			_, err := client.Write([]byte("FILE_TRANSFER_TO_CLI_START\n"))
			if err != nil {
				updateOutput(resultOutput, fmt.Sprintf("通知客户端失败: %v\n", err))
				continue
			}
			fileName := reader.URI().Name()
			_, err = client.Write([]byte(fileName + "\n"))
			if err != nil {
				updateOutput(resultOutput, fmt.Sprintf("发送文件名失败: %v\n", err))
				continue
			}

			fileSize := len(fileContent)
			_, err = client.Write([]byte(fmt.Sprintf("%d\n", fileSize)))
			if err != nil {
				updateOutput(resultOutput, fmt.Sprintf("发送文件大小失败: %v\n", err))
				continue
			}
			_, err = client.Write(fileContent)
			if err != nil {
				updateOutput(resultOutput, fmt.Sprintf("发送文件失败: %v\n", err))
			}
		}
	}, w)
}
