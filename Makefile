prep:
	./scripts/prep.sh $(n)

stlog:
	./scripts/start_with_logs.sh 

stmux:
	./scripts/start_with_tmux.sh 

clean:
	./scripts/clean_up.sh

reset-db:
	./scripts/reset_db.sh

multi-logs:
	# Display logs with multitail
	multitail -s 2 -sn 1,2 cache/log.txt \
						balancer/log.txt \
						receiver/node0/log.txt