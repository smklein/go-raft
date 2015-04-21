source ./sourceme.sh
./build.sh
cd tests

tests=( "basic_leader_test.go" "basic_write_test.go" \
        "sequential_write_test.go" "concurrent_write_test.go" \
        "basic_persistent_log_test.go" )

for t in "${tests[@]}"
do
    printf ">>> TESTING $t ..."
    rm -f output.txt
    go test -v $t > output.txt
    fail=$(grep "FAIL" output.txt | wc -l)
    if [ $fail != 0 ]; then
        printf "... FAILED $t. Check output.txt for results <<< ✗\n"
        exit 1
    fi
    printf "... PASSED $t <<< ✓\n"
done

cd ..
