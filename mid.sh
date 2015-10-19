#!/bin/bash
stop(){
if [ `ps aux | grep bloom80 | grep -v grep | wc -l` -ge 1 ];then
	kill -9 `ps aux | grep bloom80 | grep -v grep | awk '{print $2}'`
	echo " bloom80 stop sucessfull"
fi
}
start(){
if [ `ps aux | grep bloom80 | grep -v grep | wc -l` == 0 ];then
	if [ `whoami` == 'root' ];then
		su - search -c '/home/s/apps/tantan_mid/antispam/bloom80 -f /home/s/data/allMid150803.log -p 8083 >> /home/s/logs/bloom.log 2>&1 &'
		[ `ps aux | grep bloom80 | grep -v grep | wc -l` == 1 ] && echo "bloom80 start sucessfull" || echo "bloom80 start fail"
	else
		/home/s/apps/tantan_mid/antispam/bloom80 -f /home/s/data/allMid150803.log -p 8083 >> /home/s/logs/bloom.log 2>&1 &
		[ `ps aux | grep bloom80 | grep -v grep | wc -l` == 1 ] && echo "bloom80 start sucessfull" || echo "bloom80 start fail"
	fi
else
	echo "bloom80 already running"
fi
}
restart(){
if [ `ps aux | grep bloom80 | grep -v grep | wc -l` == 1 ];then
	stop
	sleep 2;
	start
	
fi
}
status(){
	[ `ps aux | grep bloom80 | grep -v grep | wc -l` == 1 ] && echo "bloom80 starting" || echo "bloom80 stoping"
}
case "$1" in
	start)
	start
	;;
	stop)
	stop
	;;
	restart)
	restart
	;;
	status)
	status
	;;
	*)
	echo $"Usage: $0 {start|stop|restart|status}"
esac
