# Dokku VM Stats plugin
======================================

Collects stats from the machine like memory, CPU usages and post requests to client service that will be listening to it.

It adds a cron job in the host machine, then runs it every minute.


Project: https://github.com/dokku/dokku

Requirements
------------
* Dokku version `0.4.0` or higher

Installation
-----------
```
# dokku 0.4.x
dokku plugin:install https://github.com/evzpav/dokku-vm-stats.git
```

## License

MIT
