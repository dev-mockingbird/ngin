<!--
 Copyright (c) 2023 Yang,Zhong
 
 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

## NGIN, not X

this is an programmable API gateway, use the following script to config your API gateway

```

# listen tls, please file the below line
cert-file = "";
# listen tls, please file the below line
key-file = "";

listen 6000 {
    # generate a request-id
    header.request-id == null {
        header.request-id = uuid;
    }
    # attach request-id to response
    response.header.request-id = header.request-id;

    host == 127.0.0.1:6000 | "hello.com" | "world.com" {
        # load balancer backend
        backend http://127.0.0.1:6090 | http://127.0.0.1:6091;

        path !~ /apis/login | /apis/register | /apis/idinfo {
            header.Authorization ~ .* {
                # authentication
                call [ POST http://127.0.0.1:6080/apis/auth ];
                response.code == 200 {
                    # forward
                    call;
                }
                return;
            }
        }
        # forward to backend
        call;
    }
}

# for another system
listen 7000 {
    call http://127.0.0.1:5050;
}

```

## FEATURE
