## molly_mysql_canal

### sync mysql data to redis in mysql binlog format

## Quick Start

### [中文文档](README.zh-cn.md)

### require

### mysql needs to enable the following configuration.
#### vim /etc/my.cnf
```ini
log-bin=mysql-bin           # Enable MySQL binary log
binlog_format=ROW           # Set the binary log format to ROW
server_id=1                 # server_id must be unique
```

### Verify the configuration
```sql
SHOW VARIABLES LIKE 'binlog_format';

SHOW VARIABLES LIKE 'log_bin';
```

### create the canal user and assign permissions
```sql
CREATE USER canal IDENTIFIED BY 'canal';
GRANT RELOAD, SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'canal'@'%';
FLUSH PRIVILEGES;
```

### create config.yaml 
```yaml
appName: molly-mysql-canal

mysql:
  addr: 127.0.0.1:3306
  username: canal
  password: canal
  serverId: 88

## Add masterName: to indicate Sentinel mode
## Multiple addrs: represents cluster mode
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
      #expression： 
      #t_user table in all databases:                                           .*\.t_user
      #all tables in the canal database:                                        canal\..*
      #databases starting with molly_, all t_user tables:                       molly_.*\.t_user
      #If there are t_user and t_user_info tables in the database canal, 
      #but you only want to sync the t_user table:                              canal.t_user\b
      tableRegex: molly_.*\.cms_device

      #initialize data from the database
      initData: true

      #clear previous data, only supports redis
      clearBeforeData: true

      #serialization method [msgpack、json、yaml、protobuf] default: json
      serializationFormat: json

      #custom primary key field. Get the first primary key in the table by default.
      #customPKColumn: id

      #If there is only one field included, only the value is saved, not the field. 
      #Try not to add includeColumnNames and excludeColumnNames at the same time.
      includeColumnNames:
        - username

      #fields to exclude
      excludeColumnNames:
        - create_time
        - create_user
        - last_update_time
        - last_update_user

      #table field name conversion, for example: last_update_time
      #lowerCamelCase: lastUpdateTime
      #upperCamelCase: LastUpdateTime
      #default: last_update_time
      fieldNameFormat: lowerCamelCase # lowerCamelCase、upperCamelCase、default

      #sync destination，[redis、console]
      syncTarget: redis

      #sync to redis
      redisRule:

        #redis key name
        keyName: cms_device

        #redis key type [string、hash]  default: string
        keyType: hash

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

### [download binary](https://github.com/thousmile/molly_mysql_canal/releases)
### config.yaml and molly_mysql_canal are in the same directory

#### Linux or MacOS
```shell
sudo chmod a+x molly_mysql_canal
./molly_mysql_canal 
```

#### Windows
```shell
./molly_mysql_canal.exe
```
