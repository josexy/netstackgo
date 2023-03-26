@REM route delete 1.0.0.0/8
@REM route delete 2.0.0.0/7
@REM route delete 4.0.0.0/6
@REM route delete 8.0.0.0/5
@REM route delete 16.0.0.0/4
@REM route delete 32.0.0.0/3
@REM route delete 64.0.0.0/2
@REM route delete 128.0.0.0/1

@REM netsh interface ipv4 add route 1.0.0.0/8 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 2.0.0.0/7 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 4.0.0.0/6 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 8.0.0.0/5 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 16.0.0.0/4 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 32.0.0.0/3 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 64.0.0.0/2 "WinTun" 192.18.0.1 metric=5 store=active
@REM netsh interface ipv4 add route 128.0.0.0/1 "WinTun" 192.18.0.1 metric=5 store=active

@REM netsh interface ipv4 show route