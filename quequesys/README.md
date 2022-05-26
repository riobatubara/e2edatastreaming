# NSQ
## A realtime distributed messaging platform

[![NSQ](https://camo.githubusercontent.com/35df65972dd10241edb2bdbd1f49f7f52b83f909b32d91f76aa6bd0c6b976ea5/68747470733a2f2f6e73712e696f2f7374617469632f696d672f6e73715f626c75652e706e67)](https://nsq.io/overview/design.html)

## INSTALLING
### Building From Source
#### Pre-requisites
 - `golang (version 1.13+ is required)`

#### Compiling
NSQ uses go modules to produce reliable builds.
```sh
git clone https://github.com/nsqio/nsq
cd nsq
make
```

#### Testing
```sh
./test.sh
```

#### Run NSQ
```sh
nsqlookupd
```

```sh
nsqd --lookupd-tcp-address=127.0.0.1:4160
```

NOTE: if your system hostname does not resolve to 127.0.0.1 then add --broadcast-address=127.0.0.1

#### Run NSQ Admin
```sh
nsqadmin --lookupd-http-address=127.0.0.1:4161
```

#### Testing NSQ
```sh
curl -d 'hello world 1' 'http://127.0.0.1:4151/pub?topic=test'
```

To verify things worked as expected, in a web browser open http://127.0.0.1:4171/ to view the nsqadmin UI and see statistics.