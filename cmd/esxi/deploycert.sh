#!/usr/bin/env sh

SERVER=https://cert.example.com
DOMAIN=esxi.example.com
BASEURL=$SERVER/cert/$DOMAIN

VARDIR=/var/certcli

NOW=$(date +%s)
NEXT=$(cat $VARDIR/next 2>/dev/null ) || NEXT=0

if [ "$NOW" -lt "$NEXT" ];
then
  NEXTSTR=$(date -D %s -d $NEXT)
  printf 'Skip, Next renewal time is: %s\n' "$NEXTSTR"
  exit
fi

printf "Refresh Time\n"
SERIAL=$(cat $VARDIR/serial 2>/dev/null ) || SERIAL="0"
NEWSERIAL=$(wget $BASEURL/serial -q -O -) || exit 1

if [ "$SERIAL" = "$NEWSERIAL" ];
then
  printf "Skip, Same serial try again later\n"
  exit
fi
printf "Not same serial!\n"

NEWNEXT=$(wget $BASEURL/next -q -O -) || exit 1

wget $BASEURL/fullchain -q -O $VARDIR/rui.crt || exit 1
mv $VARDIR/rui.crt /etc/vmware/ssl/rui.crt
printf '%s' "$NEWNEXT" > $VARDIR/next
printf '%s' "$NEWSERIAL" > $VARDIR/serial
/etc/init.d/hostd restart
/etc/init.d/vpxa restart


# Edit /var/spool/cron/crontabs/root
# Add the line (all on one line) 5 0 * * * /bin/deploycert.sh >> /vmfs/volumes/Internal/cert.log 2>&1
# Run the command `kill $(cat /var/run/crond.pid)` to kill the currently running cron daemon.
# Restart cron with the command `crond`
#
# In /etc/rc.local.s/local.sh
# /bin/kill $(cat /var/run/crond.pid)
# /bin/echo '5 0 * * * /bin/deploycert.sh >> /vmfs/volumes/Internal/cert.log 2>&amp;1' >> /var/spool/cron/crontabs/root
# /bin/crond
