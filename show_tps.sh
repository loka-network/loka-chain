if [ "$1" = "-a" ]; then
    stdbuf -oL tail -f nohup.out | grep -E --line-buffered "block_size|tps|num_txs|num_txs_res" | awk '
    {
        # Remove ANSI color codes
        gsub(/\033\[[0-9;]*[mGK]/, "")
        
        # Check and print lines where block_size > 0
        if (match($0, /block_size=[0-9]+/)) {
            split(substr($0, RSTART, RLENGTH), a, "=")
            if (a[2]+0 > 0) {print; next}
        }
        # Check and print lines where tps > 0
        if (match($0, /tps=[0-9]+/)) {
            split(substr($0, RSTART, RLENGTH), a, "=")
            if (a[2]+0 > 0) {print; next}
        }
        # Check and print lines where num_txs > 0
        if (match($0, /num_txs=[0-9]+/)) {
            split(substr($0, RSTART, RLENGTH), a, "=")
            if (a[2]+0 > 0) {print; next}
        }
        # Check and print lines where num_txs_res > 0
        if (match($0, /num_txs_res=[0-9]+/)) {
            split(substr($0, RSTART, RLENGTH), a, "=")
            if (a[2]+0 > 0) {print; next}
        }
    }'
else
    stdbuf -oL tail -f nohup.out | grep -E --line-buffered "tps" | awk '
    {
        # Remove ANSI color codes
        gsub(/\033\[[0-9;]*[mGK]/, "")
        
        # Check and print lines where tps > 0
        if (match($0, /tps=[0-9]+/)) {
            split(substr($0, RSTART, RLENGTH), a, "=")
            if (a[2]+0 > 0) {print; next}
        }
    }'
fi