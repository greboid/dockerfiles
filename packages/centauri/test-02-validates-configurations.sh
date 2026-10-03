#!/bin/sh
printf 'route test.example.internal\n\tupstream 127.0.0.1:9999\n' > /tmp/good.conf
/usr/bin/centauri -config /tmp/good.conf -validate
printf 'not a centauri config\n' > /tmp/bad.conf
! /usr/bin/centauri -config /tmp/bad.conf -validate
