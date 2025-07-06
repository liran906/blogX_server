-- 这个文件只是记录在远程主机的操作
-- 因为我的版本是 8.0; 和教程中的 5.7 有一些差异

-- 安装 配置好 mysql 的 docker 之后

cd /opt/blogx/blogX_server/init/deploy

docker compose up -d

-- 跑起来
-- 进主节点的mysql

docker exec -it mysql-master bash

mysql -uroot -p123456

docker exec -it mysql-slave bash

-- ============================
-- 💻 在主库 (10.2.0.2:3406) 执行
-- ============================

-- 1️⃣ 创建一个专门用于主从复制的用户 'repl'，密码为 '123456'，允许所有 IP 访问（@'%' 表示任意 IP）
CREATE USER 'repl'@'%' IDENTIFIED BY '123456';

-- 2️⃣ 给该用户授予主从复制权限
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';

-- 3️⃣ 刷新权限，让前面的操作生效
FLUSH PRIVILEGES;

-- 4️⃣ 查看当前主库的二进制日志文件名和位置（用于从库连接同步时指定）
SHOW MASTER STATUS \G;

-- 📌 结果类似：
# *************************** 1. row ***************************
#              File: mysql-bin.000003
#          Position: 827
#      Binlog_Do_DB:
#  Binlog_Ignore_DB:
# Executed_Gtid_Set:
# 1 row in set (0.01 sec)

-- 记下 `File` 和 `Position`，稍后配置从库需要用到

-- ============================
-- 🗄️ 在从库 (假设 10.2.0.3) 执行
-- ============================

-- 1️⃣ 配置主库信息及同步位置
STOP REPLICA IO_THREAD FOR CHANNEL ''; -- 如果是第二次以上运行

CHANGE MASTER TO
    MASTER_HOST='10.2.0.2',         -- 主库 IP
    MASTER_PORT=3306,               -- 主库端口（⚠️必须指定！）
    MASTER_USER='repl',             -- 主从同步用户
    MASTER_PASSWORD='123456',       -- 同步用户密码
    MASTER_LOG_FILE='mysql-bin.000007', -- 从主库 SHOW MASTER STATUS 得到
    MASTER_LOG_POS=1182;                 -- 从主库 SHOW MASTER STATUS 得到


CHANGE MASTER TO MASTER_HOST='10.2.0.2', MASTER_PORT=3306, MASTER_USER='repl', MASTER_PASSWORD='123456', MASTER_LOG_FILE='mysql-bin.000003', MASTER_LOG_POS=157;

-- 2️⃣ 启动从库同步线程
START SLAVE;

-- 3️⃣ 查看从库同步状态，确认是否成功
SHOW SLAVE STATUS \G;

-- 📌 确认以下字段都为 Yes
-- Slave_IO_Running: Yes
-- Slave_SQL_Running: Yes
-- 如果出现错误，可以看 Last_IO_Error 和 Last_SQL_Error 字段，排查问题