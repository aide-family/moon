INSERT INTO `prom_alarm_histories` (`deleted_at`, `instance`, `status`, `info`, `start_at`, `end_at`, `duration`,
                                    `strategy_id`, `level_id`, `md5`)
VALUES (0, 'prom:9090', 1,
        '{\"status\":\"firing\",\"labels\":{\"alertname\":\"up\",\"instance\":\"prom:9090\",\"job\":\"prometheus_server\",\"severity\":\"warning\"},\"annotations\":{\"description\":\"prom:9090 of job prometheus_server has been down for more than 1 minute.\",\"title\":\"Instance prom:9090 down\"},\"startsAt\":1697516631,\"endsAt\":-62135596800,\"generatorURL\":\"http://8ce022f7e801:9090/graph?g0.expr=up+%3D%3D+1\\u0026g0.tab=1\",\"fingerprint\":\"a2f1d053912a1d40\"}',
        1697516631, -62135596800, -9223372036, 0, 0, 'a2f1d053912a1d40'),
       (0, 'pushgateway:9091', 1,
        '{\"status\":\"firing\",\"labels\":{\"alertname\":\"up\",\"instance\":\"pushgateway:9091\",\"job\":\"pushgateway\",\"severity\":\"warning\"},\"annotations\":{\"description\":\"pushgateway:9091 of job pushgateway has been down for more than 1 minute.\",\"title\":\"Instance pushgateway:9091 down\"},\"startsAt\":1697516631,\"endsAt\":-62135596800,\"generatorURL\":\"http://8ce022f7e801:9090/graph?g0.expr=up+%3D%3D+1\\u0026g0.tab=1\",\"fingerprint\":\"dfd0a50a1843cd59\"}',
        1697516631, -62135596800, -9223372036, 0, 0, 'dfd0a50a1843cd59')
ON DUPLICATE KEY UPDATE `status`=VALUES(`status`),
                        `end_at`=VALUES(`end_at`),
                        `duration`=VALUES(`duration`),
                        `info`=VALUES(`info`)