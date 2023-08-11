# badger
![logo](media/badger.svg)

Badger is a lightweight alternative to systemd, it runs your scripts and manages your daemons at startup.

### Features
* Live reload of the daemons file allows your to add and remove daemons and scripts while system is running live
* Automatically restarts failed daemons and stops removed daemons
* Offers a simple replacement for an otherwise heavy init daemon like systemd.
* Designed to be used with runc, systemd and serve as systemd replacement in docker containers
* Properly reaps zombies
* Follows KISS principle

### initFile
```
s:/runMeOnce.sh param1
d:/iAmDaemon.sh
d:/usr/sbin/sshd -D
```

### Running
#### Using inplace of `/sbin/init`
`badger` can be used inplace of `/sbin/init`, for example when using microvms.
In those cases as no arguments are passed it will read config from `/initrc` and write log to `init.log`


#### Shell
```bash
./badger if /initFile log /log/badger.log
```

#### runC
Place badger in /bin/badger inside `container/rootfs/bin/badger` and create the `container/rootfs/initFile` and log folder.
runC config.json example:
```json
	"args": [
		"/bin/badger if /initFile log /log/badger.log"
	],
```

#### systemD
Using badger to start your services on startup
Add contents to `/etc/systemd/system/badger.service`
```
[Unit]
Description=badger
After=network.target

[Service]
ExecStart=/bin/badger if /initFile log /log/badger.log
Type=simple

[Install]
WantedBy=multi-user.target
Alias=badger.service
```

Then run:
```bash
systemctl enable badger
systemctl start badger
```

## Installing
```bash
repo="badger"; name="badger"; os=$(uname | tr '[:upper:]' '[:lower:]'); arch=$(uname -m); case $arch in x86_64) arch="amd64" ;; arm64) arch="arm64" ;; esac; url="https://github.com/8ff/${repo}/releases/download/latest/${name}.${os}.${arch}"; curl -L $url -o ${name} && chmod +x ${name}
```
