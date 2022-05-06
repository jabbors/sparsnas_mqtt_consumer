#!/bin/bash

docker-compose exec influxdb influx -execute "SHOW DATABASES" | grep "sparsnas" && exit

echo "creating database"
docker-compose exec influxdb influx -execute "CREATE DATABASE sparsnas"
docker-compose exec influxdb influx -execute "ALTER RETENTION POLICY \"autogen\" ON \"sparsnas\" DURATION 1h SHARD DURATION 30m"
docker-compose exec influxdb influx -execute "CREATE RETENTION POLICY \"day\" ON \"sparsnas\" DURATION 24h REPLICATION 1"
docker-compose exec influxdb influx -execute "CREATE RETENTION POLICY \"week\" ON \"sparsnas\" DURATION 7d REPLICATION 1"
docker-compose exec influxdb influx -execute "CREATE RETENTION POLICY \"month\" ON \"sparsnas\" DURATION INF REPLICATION 1"
docker-compose exec influxdb influx -execute "CREATE CONTINUOUS QUERY \"day_agg\" ON \"sparsnas\" BEGIN SELECT MEAN(\"watt\") AS \"watt\",MAX(\"kwh\") AS \"kwh\",MEAN(\"battery\") AS \"battery\",MEAN(\"freqerr\") AS \"freqerr\",MEAN(\"effect\") AS \"effect\" INTO \"sparsnas\".\"day\".\"reading\" FROM \"sparsnas\".\"autogen\".\"reading\" GROUP BY time(1m),sensor END"
docker-compose exec influxdb influx -execute "CREATE CONTINUOUS QUERY \"week_agg\" ON \"sparsnas\" BEGIN SELECT MEAN(\"watt\") AS \"watt\",MAX(\"kwh\") AS \"kwh\",MEAN(\"battery\") AS \"battery\",MEAN(\"freqerr\") AS \"freqerr\",MEAN(\"effect\") AS \"effect\" INTO \"sparsnas\".\"week\".\"reading\" FROM \"sparsnas\".\"day\".\"reading\" GROUP BY time(10m),sensor END"
docker-compose exec influxdb influx -execute "CREATE CONTINUOUS QUERY \"month_agg\" ON \"sparsnas\" BEGIN SELECT MEAN(\"watt\") AS \"watt\",MAX(\"kwh\") AS \"kwh\",MEAN(\"battery\") AS \"battery\",MEAN(\"freqerr\") AS \"freqerr\",MEAN(\"effect\") AS \"effect\" INTO \"sparsnas\".\"month\".\"reading\" FROM \"sparsnas\".\"week\".\"reading\" GROUP BY time(1h),sensor END"
docker-compose exec influxdb influx -execute "CREATE RETENTION POLICY \"grafana_rp\" ON \"sparsnas\" DURATION INF REPLICATION 1"
docker-compose exec influxdb influx -execute "INSERT INTO \"sparsnas\".\"grafana_rp\" config rp=\"autogen\",gb=\"15s\" 3600000000000"
docker-compose exec influxdb influx -execute "INSERT INTO \"sparsnas\".\"grafana_rp\" config rp=\"day\",gb=\"1h\" 86400000000000"
docker-compose exec influxdb influx -execute "INSERT INTO \"sparsnas\".\"grafana_rp\" config rp=\"week\",gb=\"24h\" 604800000000000"
docker-compose exec influxdb influx -execute "INSERT INTO \"sparsnas\".\"grafana_rp\" config rp=\"month\",gb=\"24h\" 9223372036854775806" #-- max ns value in a 64-bit int