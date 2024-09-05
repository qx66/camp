create database camp default charset utf8mb4;

create user camp IDENTIFIED BY "pasd8KXadK#1l2x3";
grant all on camp.* to 'camp'@'%';
flush privileges;



drop table if exists instruct;
create table if not exists instruct
(
    uuid          varchar(48) primary key comment 'uuid',
    org_uuid      varchar(48) comment 'uuid',
    group_uuid    varchar(48) comment 'uuid',
    instance_name varchar(150),
    `type`        int comment '类型',
    content       text comment '指令内容',
    result        int comment '结果，0执行中，-1失败，1成功',
    reply         longtext comment '指令返回内容',
    create_time   bigint,
    update_time   bigint,
    key type (type)
) comment '指令';


drop table if exists instance;
create table if not exists instance
(
    uuid          varchar(48) primary key comment 'uuid',
    org_uuid      varchar(48) comment '组织uuid',
    group_uuid    varchar(48) comment '组uuid',
    instance_name varchar(150) comment '实例名',
    client_ip     varchar(20) comment '客户端IP',
    create_time   bigint comment '创建时间',
    update_time   bigint comment '更新时间',
    unique key org_group_instance (org_uuid, group_uuid, instance_name)
) comment '实例';

