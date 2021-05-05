package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// User 定义`User`结构
type User struct {
	Name   string      // 名字
	Id     string      // 唯一id
	Msg    chan string // 管道
	IsQuit chan bool   // 退出信号
}

// 定义全局变量
var (
	// AllUsers 映射字典
	AllUsers = make(map[string]*User)
	// Message 全局通道
	Message = make(chan string, 100)
	// Mux 全局锁
	Mux sync.Mutex
)

func main() {
	// 创建服务器
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	fmt.Println("服务器启动成功！")

	// 启动广播协程，其中带有一个全局变量的channel，负责将收到的信息传递给所有用户
	go broadcast()

	for {
		// 监听是否有新连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			return
		}

		// 建立连接
		fmt.Printf("来自[%s]的地址建立连接成功！", conn.RemoteAddr().String())

		// 为当前连接启动协程，其中包括新建用户，并且监听用户的信息，以及是否退出等等
		go handler(conn)
	}
}

// 接收数据，进行业务处理
func handler(conn net.Conn) {
	// 创建新用户，客户端与服务器建立连接时，能获取ip和port，使用这些信息作为该User的唯一标识
	clientAddr := conn.RemoteAddr().String()
	newUser := &User{
		Name:   clientAddr, // 可以修改，会提供rename命令修改，初始值与id相同
		Id:     clientAddr, // 不可更改
		Msg:    make(chan string, 100),
		IsQuit: make(chan bool), // 初始为false
	}

	// 添加新用户到全局变量
	AllUsers[clientAddr] = newUser

	// 启动协程，负责将msg中的信息返回给客户端
	go writeBackToClient(newUser, conn)

	// 启动go程，负责监听当前用户是否退出
	go watchIsQuit(newUser, conn)

	// 向广播管道中`写入上线信息`
	loginInfo := fmt.Sprintf("广播消息：[%s]上线了！\n", clientAddr)
	Message <- loginInfo

	for {
		// 初始化一片空间用于接收客户端信息
		buf := make([]byte, 1024)

		// 读取客户端发过来的请求数据，cnt为读取的长度
		cnt, _ := conn.Read(buf)

		// ctrl+c，当前用户主动退出，将退出标志置true
		if cnt == 0 {
			newUser.IsQuit <- true
		}

		// 将收到的信息转成字符串格式
		info := string(buf[:cnt-1])

		fmt.Printf("服务器接收客户端发送过来的数据为: %s\n", info)

		if info == "\\who" {
			// 遍历所有用户取出其中的信息，并只返回给输入该命令的用户
			var userInfos []string
			userInfos = append(userInfos, "============================")
			userInfos = append(userInfos, "Below are the online users:")
			for _, user := range AllUsers {
				userInfos = append(userInfos, user.Name)
			}
			userInfos = append(userInfos, "============================\n")
			newUser.Msg <- strings.Join(userInfos, "\n")
		} else if info == "\\rename" {
			_, _ = conn.Write([]byte("请问你的新用户名是："))
			cnt, _ = conn.Read(buf)
			// 将收到的信息转成字符串格式
			newName := string(buf[:cnt-1])
			newUser.Name = newName
			_, _ = conn.Write([]byte("更改成功！\n"))
		} else if info == "\\bye" {
			// 退出命令，同样将标志置true
			newUser.IsQuit <- true
		} else {
			// 使用 info 拼接新的字符串进行广播通知
			Message <- fmt.Sprintf("[%s]: %s\n", newUser.Name, info)
		}
	}
}

// 向所有的用户广播消息
func broadcast() {
	fmt.Println("广播go程启动成功！")

	defer fmt.Println("广播go程broadcast退出！")

	for {
		// 1. 从Message中读取数据
		currInfo := <-Message

		// 2. 将数据写入到每一个用户的Msg管道中
		fmt.Print("Message接收到消息：" + currInfo)

		for _, user := range AllUsers {
			user.Msg <- currInfo
		}
	}
}

// 每个用户应该还有一个用来自身msg管道的方法
func writeBackToClient(user *User, conn net.Conn) {
	for {
		info := <-user.Msg

		_, err := conn.Write([]byte(info))
		if err != nil {
			return
		}
	}
}

// 监听用户是否退出
func watchIsQuit(user *User, conn net.Conn) {
	fmt.Printf("正在监听[%s]是否退出！\n", user.Id)

	defer fmt.Printf("用户[%s]已下线！", user.Id)

	for {
		select {
		case <-user.IsQuit:
			logInfo := fmt.Sprintf("广播消息：[%s]下线了！\n", user.Id)
			Message <- logInfo
			// 删除操作时加锁
			Mux.Lock()
			delete(AllUsers, user.Id)
			Mux.Unlock()
			_ = conn.Close()
			return
		}
	}
}
