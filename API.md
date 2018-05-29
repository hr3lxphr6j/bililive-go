# Bililive-go API

## About Token
你可以在配置中指定token来保证API的安全性，token可以以两种方式提供
   - HTTP Basic: `http://token:114514@127.0.0.1:8080/lives`
   - Url Parameter: `http://127.0.0.1:8080/lives?token=114514`

* `GET /lives` Get all live info 
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/lives
        ```
    - Response:   
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": {
                "lives": [
                    {
                        "id": "dad99e07cf99226b928143e7bd55b6e1",
                        "live_url": "https://live.bilibili.com/953650",
                        "info": {
                            "host_name": "真·凤舞九天",
                            "room_name": "尬聊",
                            "status": true
                        },
                        "is_listening": true,
                        "is_recoding": true
                    },
                    {
                        "id": "8cfc58ff74b31970899c0fe69345c222",
                        "live_url": "https://www.panda.tv/10300",
                        "info": {
                          "host_name": "司机王老菊",
                          "room_name": "【王老菊】",
                          "status": false
                        },
                        "is_listening": true,
                        "is_recoding": false
                    }
                ]
            }
        }
        ```
        
* `GET /lives/{id}` Get live info by id
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/lives/8cfc58ff74b31970899c0fe69345c222
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": {
                "id": "8cfc58ff74b31970899c0fe69345c222",
                "live_url": "https://www.panda.tv/10300",
                "info": {
                    "host_name": "司机王老菊",
                    "room_name": "【王老菊】",
                    "status": false
                },
                "is_listening": true,
                "is_recoding": false
            }
        }
        ```
        
* `POST /lives` Add live
    - Request:  
        ```text
        method: POST
        path: http://token:114514@127.0.0.1:8080/lives
        body: 
              {
                  "lives": [
                      {
                          "url": "https://www.panda.tv/10300",
                          "listen": true
                      }
                  ]
              }
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": {
                "lives": [
                    {
                        "id": "8cfc58ff74b31970899c0fe69345c222",
                        "live_url": "https://www.panda.tv/10300",
                        "info": {
                            "host_name": "司机王老菊",
                            "room_name": "【王老菊】",
                            "status": false
                        },
                        "is_listening": true,
                        "is_recoding": false
                    }
                ]
            }
        }
        ```        
        
* `DELETE /lives/{id}` Delete live by id
    - Request:  
        ```text
        method: DELETE
        path: http://token:114514@127.0.0.1:8080/lives/8cfc58ff74b31970899c0fe69345c222
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": "OK"
        }
        ```

* `GET /lives/{id}/start` Start listen live by id
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/lives/8cfc58ff74b31970899c0fe69345c222/start
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": {
                "id": "8cfc58ff74b31970899c0fe69345c222",
                "live_url": "https://www.panda.tv/10300",
                "info": {
                    "host_name": "司机王老菊",
                    "room_name": "【王老菊】",
                    "status": false
                },
                "is_listening": true,
                "is_recoding": false
            }
        }
        ```
        
* `GET /lives/{id}/stop` Stop listen and record live by id
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/lives/8cfc58ff74b31970899c0fe69345c222/stop
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": {
                "id": "8cfc58ff74b31970899c0fe69345c222",
                "live_url": "https://www.panda.tv/10300",
                "info": {
                    "host_name": "司机王老菊",
                    "room_name": "【王老菊】",
                    "status": false
                },
                "is_listening": false,
                "is_recoding": false
            }
        }
        ```
        
* `GET /config` Get config info
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/config
        ```
    - Response:
        ```json
        {
            "RPC": {
                "Enable": true,
                "Port": "127.0.0.1:8080",
                "Token": "114514",
                "TLS": {
                    "Enable": false,
                    "CertFile": "",
                    "KeyFile": ""
                }
            },
            "LogLevel": "info",
            "Interval": 15,
            "OutPutPath": "/Users/chigusa/Movies",
            "LiveRooms": [
                "https://live.bilibili.com/953650",
                "https://live.bilibili.com/146910"
            ]
        }
        ```
        
* `PUT /config` Save lives info to config file
    - Request:  
        ```text
        method: PUT
        path: http://token:114514@127.0.0.1:8080/config
        ```
    - Response:
        ```json
        {
            "err_no": 0,
            "err_msg": "",
            "data": "OK"
        }
        ```

* `GET /files/` A basic file server for out put path
    - Request:  
        ```text
        method: GET
        path: http://token:114514@127.0.0.1:8080/files
        ```
    - Response:
        ```html
        <pre>
            <a href="%5B2018-05-12%2021-47-52%5D%5B%E7%9C%9F%C2%B7%E5%87%A4%E8%88%9E%E4%B9%9D%E5%A4%A9%5D%5B%E5%B0%AC%E8%81%8A%5D.flv">[2018-05-12 21-47-52][真·凤舞九天][尬聊].flv</a>
            <a href="%5B2018-05-12%2021-54-27%5D%5B%E7%9C%9F%C2%B7%E5%87%A4%E8%88%9E%E4%B9%9D%E5%A4%A9%5D%5B%E5%B0%AC%E8%81%8A%5D.flv">[2018-05-12 21-54-27][真·凤舞九天][尬聊].flv</a>
            <a href="%5B2018-05-12%2021-56-22%5D%5B%E7%9C%9F%C2%B7%E5%87%A4%E8%88%9E%E4%B9%9D%E5%A4%A9%5D%5B%E5%B0%AC%E8%81%8A%5D.flv">[2018-05-12 21-56-22][真·凤舞九天][尬聊].flv</a>
            <a href="%5B2018-05-12%2022-19-15%5D%5B%E7%9C%9F%C2%B7%E5%87%A4%E8%88%9E%E4%B9%9D%E5%A4%A9%5D%5B%E5%B0%AC%E8%81%8A%5D.flv">[2018-05-12 22-19-15][真·凤舞九天][尬聊].flv</a>
        </pre>
        ```