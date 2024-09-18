#!/bin/sh
echo The value of CMD_PARAMS_ENV is: $CMD_PARAMS_ENV
/app/$CMD_PARAMS_ENV -c /data/conf -config_ext $CONFIG_TYPE