## molly_mysql_canal
### 必要条件

### mysql 需要开启如下配置。
#### vim /etc/my.cnf
```ini
log-bin=mysql-bin           # 开启MySQL二进制日志
binlog_format=ROW           # 将二进制日志的格式设置为ROW
server_id=1                 # server_id 必须唯一
```

### 验证 mysql 配置是否正确
```sql
# 确保 binlog_format=ROW
SHOW VARIABLES LIKE 'binlog_format';

# 确保 log_bin=ON
SHOW VARIABLES LIKE 'log_bin';
```

### 创建 canal 用户。并且赋值权限
```sql
CREATE USER canal IDENTIFIED BY 'canal';
GRANT RELOAD, SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';
FLUSH PRIVILEGES;
```

### 创建 config.yaml 
```yaml
appName: molly-mysql-canal

## mysql 用户 必须拥有 RELOAD,SELECT, REPLICATION SLAVE, REPLICATION CLIENT 的权限。缺一不可
mysql:
  addr: 127.0.0.1:3306
  username: canal
  password: canal
  serverId: 88

## 添加 masterName: 表示 Sentinel 模式
## 多个 addrs: 表示 集群模式
redis:
  addrs:
    - 192.168.0.188:6379
  db: 5
  #username: redis123
  #password: redis123456
  #masterName: mymaster
  #sentinelUsername: sentinel123456
  #sentinelPassword: sentinel123456


rules:
  - mysql_cms_device_to_redis:
      #表达式规则： 
      #所有库中的t_user表:              .*\.t_user
      #canal库中的所有表:                canal\..*
      #以molly_开头的库，t_user表:       molly_.*\.t_user
      #如果canal库中，有 t_user 和 t_user_info 但是只想同步 t_user 表: canal.t_user\b
      tableRegex: molly_.*\.cms_device

      #是否初始化数据
      initData: true

      #是否清空之前的数据，仅支持redis
      clearBeforeData: true

      #序列化方式 支持[msgpack、json、yaml、protobuf] 默认: json
      serializationFormat: json

      #自定义 主键字段。默认获取表中的第一个主键。
      #customPKColumn: id

      #包含的字段，包含的字段只有一个，那么仅保存值，不保存字段。 includeColumnNames 和 excludeColumnNames 尽量不要同时添加。
      includeColumnNames:
        - username

      #需要排除的字段
      excludeColumnNames:
        - create_time
        - create_user
        - last_update_time
        - last_update_user

      #数据库字段名称转换，例: last_update_time
      #lowerCamelCase: lastUpdateTime
      #upperCamelCase: LastUpdateTime
      #default: last_update_time
      fieldNameFormat: lowerCamelCase # lowerCamelCase、upperCamelCase、default

      #同步的目的地，redis、console
      syncTarget: redis

      #同步到redis
      redisRule:

        #redis中指定的key
        keyName: cms_device

        #redis中key的类型 string、hash
        keyType: hash   # string or hash

```

#### 提示：protobuf格式，使用google/protobuf/struct.proto作为交互格式
##### java 案例
```java
var bytes = new byte[]{};
var obj = Struct.parseFrom(bytes);
System.out.println(obj);
```

##### golang 案例
```go
var bytes []byte
var obj structpb.Struct
_ = proto.Unmarshal(bytes, &obj)
fmt.Println(obj.String())
```

```shell
docker run -d --name molly_mysql_canal -v /etc/molly_mysql_canal/config.yaml:/work/config.yaml --restart=always thousmile/molly_mysql_canal:1.0
```

vim docker-compose.yml

```yaml
services:

  molly_mysql_canal:
    image: thousmile/molly_mysql_canal:1.0
    container_name: molly_mysql_canal
    volumes:
      - /etc/molly_mysql_canal/config.yaml:/work/config.yaml
      - /etc/localtime:/etc/localtime:ro
    privileged: true
    restart: always

```

```shell
docker compose up -d molly_mysql_canal
```

### [下载可执行文件](https://github.com/thousmile/molly_mysql_canal/releases)
### 将 config.yaml 和 molly_mysql_canal 放在同一目录下

#### Linux or MacOS
```shell
# 添加执行权限
sudo chmod a+x molly_mysql_canal
# 运行
./molly_mysql_canal 
```

#### Windows
```shell
# 运行
./molly_mysql_canal.exe
```
