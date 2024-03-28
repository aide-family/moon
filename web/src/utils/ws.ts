let websocket: WebSocket,
    lockReconnect = false
let createWebSocket = (url: string) => {
    const userInfoStr = localStorage.getItem('user') || '{}'
    const userInfo = JSON.parse(userInfoStr)
    websocket = new WebSocket(`${url}?userId=${userInfo?.id || 0}`)
    websocket.onopen = function () {
        heartCheck.reset().start()
    }
    websocket.onerror = function () {
        reconnect(url)
    }
    websocket.onclose = function (e) {
        console.log(
            'websocket 断开: ' + e.code + ' ' + e.reason + ' ' + e.wasClean
        )
    }
    websocket.onmessage = function (event: MessageEvent<any>) {
        lockReconnect = true
        //event 为服务端传输的消息，在这里可以处理
        console.log('event', event)
    }
    return websocket
}
let reconnect = (url: string) => {
    if (lockReconnect) return
    //没连接上会一直重连，设置延迟避免请求过多
    setTimeout(function () {
        createWebSocket(url)
        lockReconnect = false
    }, 4000)
}

export type heartCheckType = {
    reset: () => heartCheckType
    start: () => void
    timeoutObj?: NodeJS.Timeout | number | string
    timeout: number
}

let heartCheck: heartCheckType = {
    timeout: 60000, //60秒
    reset: function () {
        if (this.timeoutObj) {
            clearInterval(this.timeoutObj)
        }
        return this
    },
    start: function () {
        this.timeoutObj = setInterval(function () {
            //这里发送一个心跳，后端收到后，返回一个心跳消息，
            //onmessage拿到返回的心跳就说明连接正常
            websocket.send('HeartBeat')
        }, this.timeout)
    }
}
//关闭连接
let closeWebSocket = () => {
    websocket && websocket.close()
}

export { websocket, createWebSocket, closeWebSocket }
