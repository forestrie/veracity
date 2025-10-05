testReplicateErrorForLogShorterThanSeal() {

  local output
  local other_tenant
  local tampered_log_url
  local tampered_seal_url
  local replicadir=$TEST_TMPDIR/merklelogs

  # Note: this tenant belongs to Joe Gough and he has promised never to fill the first massif
  other_tenant=tenant/97e90a09-8c56-40df-a4de-42fde462ef6f

  rm -rf $replicadir
  # first get the prod public tenant replicated for massif 0. NOTE: this is a full massif
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --massif 0 --replicadir=$replicadir)
  assertEquals "0: should return a 0 exit code" 0 $?

  COUNT=$(find $replicadir -type f | wc -l | tr -d ' ')
  assertEquals "should replicate one massif and one seal" "2" "$COUNT"

  # now get a different prod tenant log and seal. NOTE the log is partially full for this tenant
  tampered_log_url=${tampered_log_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifs/0000000000000000.log}
  tampered_seal_url=${tampered_seal_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifseals/0000000000000000.sth}
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $tampered_log_url -o tampered.log

  ## copy over the different (shorter) tenant log for massif 0
  cp tampered.log $replicadir/log/$PROD_PUBLIC_LOGID/massifs/0000000000000000.log
  # attempt to replicate the logs again, the local log data is for the wrong
  # tenant and is *less* than the seal expects, but the local seal is correct
  # for the replaced data and the remote seal
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --replicadir=$replicadir)
  status=$?
  assertEquals "1: a tampered log should exit 1" 1 $status
  assertContains "$output" "error: there is insufficient data in the massif context to generate a consistency proof against the provided state"
}

# test veracity can't extend the replica of the wrong tenant
#
# When extending a local replica, if the local tenant log data is from a tenant
# other than the requested remote, replication should fail due to consistency
# checks. This is essentially equivalent to a tamper attempt.
#
# There are two cases that are important:
#   1. The higest indexed local massif is incomplete and so the remote massif is used to extend it.
#   2. The highest indexed local massif is complete and so the remote massif is copied, leaving the original unchanged.
#
# This test suite covers only the second case. Thee first case can only be
# tested with synthesized data, or interaction with a live system, and so is
# easier to cover in the integration tests. However, the same checks are
# excercised in both cases and so this test gives a lot of confidence both
# situations are sound.
#
# Note that the --ancestor flag can be used to limit how many massifs are
# replicated. This can cause the replica to "start again" because the replica is
# so far behind that the --ancestor limit forces a gap. In this case consistency
# of the remote is not checked against the local massif, and in that case the
# replication would succeded, the local replica of the foregn tenant would not
# be updated. And the replica would be left with massifs from multiple tenants.
testReplicateErrorForMixedTenants() {

  local output
  local other_tenant
  local tampered_log_url
  local tampered_seal_url
  local replicadir=$TEST_TMPDIR/merklelogs

  # Note: this tenant is known to have > 1 massif at the time of writing and logs don't get shorter
  other_tenant=tenant/b197ba3c-44fe-4b1a-bbe8-bd9674b2bd17

  rm -rf $replicadir
  # first get the prod public tenant replicated for massif 0. NOTE: this is a full massif
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --massif 0 --replicadir=$replicadir)
  assertEquals "should return a 0 exit code" 0 $?

  COUNT=$(find $replicadir -type f | wc -l | tr -d ' ')
  assertEquals "should replicate one massif and one seal" "2" "$COUNT"

  # now get a different prod tenant log and seal. NOTE the log is partially full for this tenant
  tampered_log_url=${tampered_log_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifs/0000000000000000.log}
  tampered_seal_url=${tampered_seal_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifseals/0000000000000000.sth}
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $tampered_log_url -o tampered.log
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $tampered_seal_url -o tampered.sth

  # copy over the different tenant log for massif 0
  cp tampered.log $replicadir/log/$PROD_PUBLIC_LOGID/massifs/0000000000000000.log

  # attempt to replicate the logs again, the local log data is for the wrong tenant but the local seal is correct for the replaced data and the remote seal
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --replicadir=$replicadir)
  assertEquals "1: extending an inconsistent replica should exit 1" 1 $?
  assertContains "$output" "error: the seal signature verification failed: failed to verify seal for massif 0"

  # now add in the seal from the other log, so that the local log and seal are consistent and locally verifiable.
  cp tampered.sth $replicadir/log/$PROD_PUBLIC_LOGID/checkpoints/0000000000000000.sth

  # attempt to replicate the logs again
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --latest --replicadir=$replicadir)
  assertEquals "2: extending an inconsistent replica should exit 1" 1 $?
  assertContains "$output" "error: consistency check failed: the accumulator produced for the trusted base state doesn't match the root produced for the seal state fetched from the log"
}

# test veracity can't update the replica for a tenant whos log has been tampered with
#
# This test repeats testReplicateErrorForMixedTenants, but does so using the
# combination of watch | replicate-logs which permits finer control over the
# replica
testWatchReplicateErrorForTamperedLog() {

  local output
  local other_tenant
  local tampered_log_url
  local tampered_seal_url

  local replicadir=$TEST_TMPDIR/merklelogs

  # Note: this tenant is known to have > 1 massif at the time of writing and logs don't get shorter
  other_logid=b197ba3c-44fe-4b1a-bbe8-bd9674b2bd17
  other_tenant=tenant/$other_logid

  # first get the prod tenant replicated for massif 0. NOTE: this is a partially full massif
  rm -rf $replicadir
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$other_tenant watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$other_tenant replicate-logs --ancestors=0 --massif 0 --replicadir=$replicadir)
  assertEquals "watch-public should return a 0 exit code" 0 $?

  COUNT=$(find $replicadir -type f | wc -l | tr -d ' ')
  assertEquals "should replicate one massif and one seal" "2" "$COUNT"

  # now get a different prod public tenant log and seal. NOTE: this is a full massif
  tampered_log_url=${tampered_log_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${PROD_PUBLIC_TENANT_ID}/0/massifs/0000000000000000.log}
  tampered_seal_url=${tampered_seal_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${PROD_PUBLIC_TENANT_ID}/0/massifseals/0000000000000000.sth}
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $tampered_log_url -o tampered.log
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $tampered_seal_url -o tampered.sth

  # copy over the different tenant log for massif 0
  cp tampered.log $replicadir/log/$other_logid/massifs/0000000000000000.log

  # attempt to replicate the logs again
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$other_tenant watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$other_tenant replicate-logs --replicadir=$replicadir)
  assertEquals "extending an inconsistent replica should exit 1" 1 $?
  assertContains "$output" "error: the seal signature verification failed: failed to verify checkpoint for massif 0: verification error"

  # now attempt to change the seal to the tampered log seal
  cp tampered.sth $replicadir/log/$other_logid/checkpoints/0000000000000000.sth

  # attempt to replicate the logs again
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$other_tenant watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$other_tenant replicate-logs --replicadir=$replicadir)
  status=$?
  assertEquals "extending an inconsistent replica should exit 1" 1 $status
  assertContains "$output" "error: consistency check failed"
}

# this test ensures that veracity refused to work with replica directories that
# mix tenant massifs together while the consistency checks would prevent
# accidental extension of the wrong log, the failre mode would be very confusing
# and potentially alarming to the user.
testWatchReplicateErrorForMixedTenants() {

  local output

  local logid=${PROD_PUBLIC_LOGID}
  local tenant=${PROD_PUBLIC_TENANT_ID}
  # Note: this tenant is known to have > 1 massif at the time of writing and logs don't get shorter
  local other_logid='b197ba3c-44fe-4b1a-bbe8-bd9674b2bd17'
  local other_tenant="tenant/$other_logid"
  local replicadir=$TEST_TMPDIR/merklelogs
  local SHA=shasum

  rm -rf $replicadir

  # replicate massif 0 from the main tenant
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$tenant watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$tenant replicate-logs --massif=0 --replicadir=$replicadir)
  assertEquals "watch-public should return a 0 exit code" 0 $?

  # explicitly fetch a massif 0 from a different tenant and place it in the same replica directory using a different filename
  local other_log_url=${other_log_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifs/0000000000000000.log}
  local other_seal_url=${other_seal_url:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${other_tenant}/0/massifseals/0000000000000000.sth}
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $other_log_url -o $replicadir/log/$logid/massifs/other_tenant.log
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $other_seal_url -o $replicadir/log/$logid/checkpoints/other_tenant.sth

  # Now attempt to extend the replica. we chose $tenant because we know it has
  # more than one massif, so this command will always attempt to extend the
  # replica directory.
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$tenant watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$tenant replicate-logs --replicadir=$replicadir)
  assertEquals "extending a replica directory with mixed tenants should exit 1" 1 $?
  assertContains "$output" "error: consistency check failed"
}
