prep:
	./scripts/prep.sh $(n)

stlog:
	./scripts/start_with_logs.sh 

stmux:
	./scripts/start_with_tmux.sh 

clean:
	pwd
	./scripts/clean_up.sh