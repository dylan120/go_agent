#!/usr/local/oms_agent/python2.7.14/bin/python
# -*- coding: utf-8 -*-
import logging
from kombu.utils import uuid as kuuid
import msgpack
import json
import redis

log = logging.getLogger(__name__)
console = logging.StreamHandler()
console.setLevel(logging.DEBUG)
logging.basicConfig(level=logging.DEBUG)
formatter = logging.Formatter("[%(asctime)s] - %(filename)s[line:%(lineno)d] %(levelname)s : %(message)s")
console.setFormatter(formatter)
log.addHandler(console)

config = {
    "interface": "120.92.133.78",
    "ret_port": 12349,
    "transport": "zeromq",
    "worker_threads": 5,
    "sock_dir": "/tmp",
    "publish_port": 12347,
    "max_open_files": 2000,
    "worker_threads": 4,
    "pki_dir": "/tmp",
    "user": "root",
    "keysize": 2048,
    "serial": "msgpack",
    "id": "master1",
    "master_uri": "tcp://120.92.133.78:12349",
}

_redis = redis.StrictRedis(host=config.get("redis_host", "127.0.0.1"), port=config.get("redis_port", 6666),
                           password=config.get("redis_passwd", "123456"))

# with open("/tmp/.root_key", "r") as fd:
# with open("/usr/local/oms_agent/pki/web/.root_key", "r") as fd:
#     key = fd.read()

for i in range(1):
    task_instance_id = kuuid()
    payload = {
        "data": {
            "project_id": 62,
            "task_instance_id": task_instance_id,
            "task_id": 0,
            "operator": "root",
            "type": "edit",
            "task_name": "腾讯云上海二区星玩、瑞玩、游爱兄弟游戏服更新版本20170715（g3xx）",
            "msg": {
                "msgType": 0
            },
            "data": {
                "blocks": [
                    {
                        "type": 1,
                        "blockOrd": 1,
                        "block_name": "腾讯云上海二区内网同步服务端包20170715",
                        "steps": [{
                            "function": "cmd.run",

                            "script_type": "2",
                            "block_name": "腾讯云上海二区内网同步服务端包20170715",
                            "is_finished": False,
                            "text": "",
                            "script_id": 237,
                            "step_id": "407eb9244f6211e7a26d5254004f1906_1_1",
                            "is_pause": False,
                            "ord": 1,
                            "account": "root",
                            "name": "腾讯云上海二区内网同步服务端包20170715",
                            "creater": "liuweijian",
                            "script_param": "",
                            "script_content": "sleep 10",
                            "script_name": "腾讯云上海二区内网同步服务端包20170715",
                            "block_ord": 2,
                            "timeout": 1200,
                            # "ip_minionid": {"121.201.6.10": "minion1", "118.89.59.120": "minion2"},
                            "minions": ["5a1cf523-39b2-4eee-8a28-978253224b70"],
                            "type": 1,
                            "timeout": 60,
                        }]
                    }, {
                        "type": 2,
                        "blockOrd": 2,
                        "block_name": "分发游戏服务端",
                        "steps": [{
                            "function": "bt.maketorrent",
                            "is_finished": False,
                            "script_type": "2",
                            "block_name": "分发游戏服务端",
                            "text": "",
                            "script_id": 237,
                            "step_id": "407eb9244f6211e7a26d5254004f1906_1_1",
                            "is_pause": False,
                            "ord": 1,
                            "account": "root",
                            "name": "腾讯云上海二区内网同步服务端包20170715",
                            "creater": "liuweijian",
                            "script_param": "",
                            "script_content": "",
                            "script_name": "腾讯云上海二区内网同步服务端包20170715",
                            "block_ord": 2,
                            # "timeout": 1200,
                            # "ip_minionid": {"121.201.1.20": "1", "120.55.163.200": "5","211.159.219.217":"6"},
                            # "ip_minionid": {"118.89.59.120": "minion2"},
                            # "minions": ["minion1", "minion4", "minion5", "minion6"],
                            "minions": ["5a1cf523-39b2-4eee-8a28-978253224b70"],
                            "type": 2,
                            "file_target_path": "/tmp/test/abc",
                            "file_source": [
                                {
                                    "account": "root",
                                    "ip_list": [
                                        "252ea5f8-d4b9-445f-8dea-fb2a1bc9b4e0"
                                    ],
                                    # "file": "/usr/local/oms_agent/oms-master,/root/mysql_privilege_reset.sh",
                                    "file": "/tmp/insert_active.sql"
                                }
                            ],
                            "timeout": 60,
                        }
                        ],
                    },
                    {
                        "type": 3,
                        "steps": [
                            {
                                "function": "mysql.query",
                                "script_type": "2",
                                "block_name": "步骤1",
                                "is_finished": False,
                                "text": "",
                                "script_id": 233,
                                "step_id": "407eb9244f6211e7a26d5254004f1906_2_1",
                                "is_pause": False,
                                "ord": 1,
                                "account": "root",
                                "name": "步骤1",
                                "creater": "liuweijian",
                                "script_param": "",
                                "sql_txt": "use mytest;select * from step_record;",
                                "mysql": {
                                    "host": "127.0.0.1",
                                    "user": "root",
                                    "passwd": "",
                                    "db": "mysql",
                                    "port": 3306,
                                },
                                "script_content": "",
                                "script_name": "步骤1",
                                "block_ord": 3,
                                "timeout": 2000,
                                "minions": ["5a1cf523-39b2-4eee-8a28-978253224b70"],
                                "type": 1,
                                "timeout": 60,
                            }
                        ],
                        "blockOrd": 3,
                        "block_name": "步骤1"
                    },
                    {
                        "type": 1,
                        "steps": [
                            {
                                "function": "cmd.run",
                                "script_type": "2",
                                "block_name": "步骤2",
                                "is_finished": False,
                                "text": "",
                                "script_id": 236,
                                "step_id": "407eb9244f6211e7a26d5254004f1906_3_1",
                                "is_pause": False,
                                "ord": 1,
                                "account": "root",
                                "name": "步骤2",
                                "creater": "liuweijian",
                                "script_param": "all",
                                "script_content": "ls;echo hi",
                                "script_name": "步骤2",
                                "block_ord": 4,
                                "timeout": 5400,
                                "minions": ["5a1cf523-39b2-4eee-8a28-978253224b70", "minion4"],
                                "type": 1,
                                "timeout": 60,
                            }
                        ],
                        "blockOrd": 4,
                        "block_name": "步骤2"
                    },
                    {
                        "type": 1,
                        "steps": [
                            {
                                "function": "cmd.run",
                                "script_type": "2",
                                "block_name": "步骤3",
                                "is_finished": False,
                                "text": "",
                                "script_id": 234,
                                "step_id": "407eb9244f6211e7a26d5254004f1906_4_1",
                                "is_pause": False,
                                "ord": 1,
                                "account": "root",
                                "name": "步骤3",
                                "creater": "liuweijian",
                                "script_param": "all",
                                "script_content": "free",
                                "script_name": "步骤3",
                                "blockOrd": 5,
                                "timeout": 200,
                                "minions": ["5a1cf523-39b2-4eee-8a28-978253224b70"],
                                "type": 1,
                                "timeout": 60,
                            }
                        ],
                        "blockOrd": 5,
                        "block_name": "步骤3"
                    }
                ]
            },
            "is_schedule": False
        },
        "crypt": "clear"
    }
    _redis.lpush("oms_jobs", json.dumps(payload))

