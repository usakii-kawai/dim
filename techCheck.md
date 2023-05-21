# 分布式唯一ID 
​    全局唯一， 写入数据库前确认，避免写入上锁或冲突
​    设计前期就使用分布式ID作为数据的主键，成本实际上是非常低的
​    数据库主键冲突后迁移成本高

    核心场景 消息存储
    
    - 有序性：提高数据库插入性能。利于数据库批处理IO，索引值越小,单位id内存越小，所以只考虑数字
    - 可用性：可保证高并发下的可用性。
    - 友好性：尽量简单易用，不增加系统负担。
    
    选型：
    数据库自增id 起点不同 步长相同 重构修改代价高
    分布式id服务
    雪花算法     需要保证每个计算节点编号NodeID不重复

!(./assets/image-20230519174448464.png)

![image-20230519174504301](./assets/image-20230519174504301.png)



# Server

![image-20230519174921580](./assets/image-20230519174921580.png)

![channel.png](https://p9-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/19590d61c5c245b9939327fd38a59020~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![image-20230519185523887](./assets/image-20230519185523887.png)

```go
type Server interface {
    SetAcceptor(Acceptor)
    SetMessageListener(MessageListener)
    SetStateListener(StateListener)
    SetReadWait(time.Duration)
    SetChannelMap(ChannelMap)
    
    Start() error
    Push(string, []byte) error
    Shutdown(context.Context) error
}
```

```go
// server.go
type Acceptor interface {
    Accept(Conn, time.Duration) (string, error)
}
```

``` go
// server.go
type StateListener interface {
    Disconnect（string) error
}
```

```go
// server.go
// heart ping-pong
type SetReadWaot(time.Duation)
```

```go
// channels.go
type ChannelMap interface {
    Add(channel Channel)
    Remove(id string)
    Get(id string) (Channel, bool)
    All() []Channel
}
```

```go
// server.go
type MessageListener interface {
    Receive(Agent, []byte)
}

type Agent interface {
    ID() string
    Push([]byte) error
}
```



# Protocol

![image-20230519180113330](./assets/image-20230519180113330.png)

```
// server.go
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}
```

```
// server.go
// Conn Connection
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}
```

```go
// server.go
const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)
```



# Client

```go
// 
type Client interface {
    ID() string
    Name() string
    Connect(string) error
    SetDialer(Dialer)
    Send([]byte) error
    Read() (Frame, error)
    Close()
}

type Dialer interface {
    DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
    Id string
    Name string
    Address string
    Timeout time.Duration
}
```

# 协议区分

```go
type Magic [4]byte

var (
    //逻辑协议
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
    //基础协议
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65} 
)

func Read(r io.Reader) (interface{}, error) {
	magic := wire.Magic{}
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}
	switch magic {
	case wire.MagicLogicPkt:
		p := new(LogicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	case wire.MagicBasicPkt:
		p := new(BasicPkt)
		if err := p.Decode(r); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, errors.New("magic code is incorrect")
	}
}
```

# container

![image-20230521114902821](./assets/image-20230521114902821.png)

![image-20230521115419949](./assets/image-20230521115419949.png)

![image-20230521115527711](./assets/image-20230521115527711.png)

# services

![image-20230521115806819](./assets/image-20230521115806819.png)

![image-20230521120057093](./assets/image-20230521120057093.png)

通信服务tcp协议服务无法感知更加抽象的http服务，也不能使用nginx等http做反向代理，所以需要支持dns的注册中心 保证cap中的c 一致性

![image-20230521120352935](./assets/image-20230521120352935.png)

service  register self by service_find dns SRV

router find service by register_center

![image-20230521121133047](./assets/image-20230521121133047.png)

逻辑服务必须与全部网关建立连接后，才能接受网关转发过来的消息, 暂时阻塞逻辑服务保证一致性

![image-20230521121405644](./assets/image-20230521121405644.png)

![image-20230521121615895](./assets/image-20230521121615895.png)

# router

![image-20230521121838333](./assets/image-20230521121838333.png)

![image-20230521122209580](./assets/image-20230521122209580.png)

![image-20230521122218976](./assets/image-20230521122218976.png)

![image-20230521122432669](./assets/image-20230521122432669.png)

责任链模式

gateway init

![image-20230521122538329](./assets/image-20230521122538329.png)

server init

![image-20230521122615288](./assets/image-20230521122615288.png)

# login & offline

![image-20230521122844393](./assets/image-20230521122844393.png)

![image-20230521122930792](./assets/image-20230521122930792.png)

# message

![image-20230521123046908](./assets/image-20230521123046908.png)

![image-20230521123101668](./assets/image-20230521123101668.png)

![image-20230521123120698](./assets/image-20230521123120698.png)

![image-20230521123135185](./assets/image-20230521123135185.png)

![image-20230521123157453](./assets/image-20230521123157453.png)
