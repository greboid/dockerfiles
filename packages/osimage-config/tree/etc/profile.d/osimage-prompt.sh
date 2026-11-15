case $- in *i*) ;; *) return 0 ;; esac

__osimage_ps1() {
	_s=$1
	[ "$2" = "$3" ] && _s=0
	_b=$(git symbolic-ref --short -q HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null) && printf ' %s' "$_b"
	if [ "$_s" -eq 0 ]; then printf '\\[\\e[32m\\]'; else printf '\\[\\e[31m\\]'; fi
}

if [ "$(id -u)" -eq 0 ]; then _c=31; else _c=32; fi
PS1='\[\e['$_c'm\]\u\[\e[0m\]@\h$(__osimage_ps1 $? $LINENO ${__osimage_line:-})❯\[\e[$(( (__osimage_line=LINENO)*0 ))m\] '
unset _c
