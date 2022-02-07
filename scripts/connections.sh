sudo netstat -tn | awk '$5 ~ /127.0.0.1:808./ { print $(NF-1) " " $(NF) }' | sort | uniq -c
