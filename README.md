# 聊天室

作为一个聊天室，里面应该有多个用户在进行聊天，所以需要定义
`User`结构。

同时，服务端接收到某个用户的发来的消息后，它应该展示给所有用户，
即进行全局广播。

同时支持以下3个功能：
- who: 查询当前在线人数
- rename: 改名
- bye: 退出

# demo

![demo](./chatroom_demo.gif)

# 后续

更完备的功能应当包含一个数据库，保存所有user的信息。还有当有一个前端，
使得用户输入的信息与收到的信息不杂糅在一起，参考QQ的前端设置。

# Reference

- [p52-p64](https://www.bilibili.com/video/BV1ZU4y1W72B)