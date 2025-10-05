testVeracityWatchPublicFindsActivity() {
  local output
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID watch --horizon 10000h)
  assertEquals "watch-public should return a 0 exit code" 0 $?
  assertContains "watch-public should find activity" "$output" "$PROD_PUBLIC_TENANT_ID"
}

testVeracityWatchLatestFindsActivity() {
  local output
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID watch --latest)
  assertEquals "watch-public --latest should return a 0 exit code" 0 $?
  assertContains "watch-public --latest should find activity" "$output" "$PROD_PUBLIC_TENANT_ID"
}

testVeracityReplicateLogsPublicTenantWatchPipe() {
  local output
  local replicadir=$TEST_TMPDIR/merklelogs

  rm -rf $replicadir
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata \
    --tenant=$PROD_PUBLIC_TENANT_ID watch --horizon 10000h |
    $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --ancestors=0 --replicadir=$replicadir)
  assertEquals "watch-public should return a 0 exit code" 0 $?

  COUNT=$(find $replicadir -type f | wc -l | tr -d ' ')
  assertEquals "should replicate one massif and one seal" "2" "$COUNT"
}

testVeracityReplicateLogsPublicTenantWatchLatestFlag() {
  local output
  local replicadir=$TEST_TMPDIR/merklelogs

  rm -rf $replicadir
  output=$($VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID replicate-logs --latest --ancestors=0 --replicadir=$replicadir)
  assertEquals "replicate-logs --latest should return a 0 exit code" 0 $?

  COUNT=$(find $replicadir -type f | wc -l | tr -d ' ')
  assertEquals "should replicate one massif and one seal" "2" "$COUNT"
}
