#! /bin/bash

VERACITY_INSTALL=${VERACITY_INSTALL:-../../veracity}
DATATRAILS_URL=${DATATRAILS_URL:-https://app.datatrails.ai}
PUBLIC_ASSET_ID=${PUBLIC_ASSET_ID:-publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8}
PUBLIC_EVENT_ID=${PUBLIC_EVENT_ID:-publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8/events/a022f458-8e55-4d63-a200-4172a42fc2aa}

PROD_PUBLIC_TENANT_ID=${PROD_PUBLIC_TENANT_ID:-tenant/6ea5cd00-c711-3649-6914-7b125928bbb4}

PROD_LOG_URL=${PROD_LOG_URL:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${PROD_PUBLIC_TENANT_ID}/0/massifs/0000000000000000.log}
TEST_TMPDIR=${TEST_TMPDIR:-${SHUNIT_TMPDIR}}
EMPTY_DIR=$TEST_TMPDIR/empty
PROD_DIR=$TEST_TMPDIR/prod
DUP_DIR=$TEST_TMPDIR/duplicate-massifs
PROD_LOCAL_BLOB_FILE="$PROD_DIR/mmr.log"
INVALID_BLOB_FILE="$TEST_TMPDIR/invalid.log"

oneTimeSetUp() {
    mkdir -p $EMPTY_DIR
    mkdir -p $PROD_DIR
    mkdir -p $DUP_DIR
    curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $PROD_LOG_URL -o $PROD_LOCAL_BLOB_FILE
    touch $INVALID_BLOB_FILE

    # Duplicate the prod and soak massif files in a single directory. The
    # replication should refuse to work with a directory that has multiple
    # massif files for the same massif index.
    cp $PROD_LOCAL_BLOB_FILE $DUP_DIR/prod-mmr.log

    assertTrue "prod MMR blob file should be present" "[ -r $PROD_LOCAL_BLOB_FILE ]"
    assertTrue "invalid MMR blob file should be present" "[ -r $INVALID_BLOB_FILE ]"
}

# tests that the replica is extended if a new entry is added to the remote log
# same as the replicate-logs-latest-interactive.sh testReplicateLatest but uses the watch command.
# Use TENANT to override the default which is the PROD_PUBLIC_TENANT_ID
# To cause the log of the prod public tenant use the UI to create a public event.
testWatchReplicateLatest() {

    local output

    local tenant=${TENANT:-$PROD_PUBLIC_TENANT_ID}
    local replicadir=$TEST_TMPDIR/merklelogs
    local SHA=shasum

    rm -rf $replicadir

    # replicate the most recent massif (--ancestors=0 assures this)
    output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
        --tenant=$tenant watch --horizon 10000h \
        | $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$tenant replicate-logs --ancestors=0 --replicadir=$replicadir)
    assertEquals "watch-public should return a 0 exit code" 0 $?

    # identify the filename of the last massif
    local last_massif=$(ls $replicadir/$tenant/0/massifs/*.log | sort -n | tail -n 1)
    echo "last_massif: $last_massif"
    local sum_first=$(shasum $last_massif | awk '{print $1}')
    echo "sum_first: $sum_first"

    while true; do
        echo "waiting for the log for $tenant to grow"
        # now get a different prod public tenant log and seal. NOTE: this is a full massif
        output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
            --tenant=$tenant watch --horizon 10000h \
            | $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$tenant replicate-logs --replicadir=$replicadir)
        echo "output: $output"

        # This handles the case where the initial last massif is perfectly full
        # at the start of the test. When this happens it does not impact the
        # validity of the test. It is not possible to test both cases in
        # isolation without creating synthetic log data in the forestrie
        # instance under test.
        # for test fail reporting we just need to know if this case was in play
        # when a failure was detected.
        local new_last_massif=$(ls $replicadir/$tenant/0/massifs/*.log | sort -n |  tail -n 1)
        if [ "$last_massif" != "$new_last_massif" ]; then
            echo "*** new event created new massif ***"
            echo "*** this is fine, but if this test fails, please include this info in the bug report"
            echo "massifa: $last_massif"
            echo "massifb: $new_last_massif"
            last_massif=$new_last_massif
        fi

        local sum_cur=$($SHA $last_massif | awk '{print $1}')
        echo "sum_first, sum_cur: $sum_first, $sum_cur"
        if [ "$sum_first" != "$sum_cur" ]; then
            echo "the log grew"
            break
        fi
        sleep 3
    done;
    # now get a different prod public tenant log and seal. NOTE: this is a full massif
}

assertStringMatch() {
    local message="$1"
    local expected="$2"
    local actual="$3"

    # Normalize by converting all spaces to a single space, removing leading/trailing spaces and punctuation.
    expected=$(echo "$expected" | sed -e 's/[[:space:]]\+/ /g' -e 's/^[[:space:]]*//;s/[[:space:]]*[[:punct:]]*$//')
    actual=$(echo "$actual" | sed -e 's/[[:space:]]\+/ /g' -e 's/^[[:space:]]*//;s/[[:space:]]*[[:punct:]]*$//')

    echo "Expected (hex):" && echo "$expected" | hexdump -C
    echo "Actual (hex):" && echo "$actual" | hexdump -C


    assertEquals "$message" "$expected" "$actual"
}
