#!/bin/sh

#--- Start the apiscout API server ---
./tmp/server &

#--- Start NGINX ---
nginx -g "daemon off;"