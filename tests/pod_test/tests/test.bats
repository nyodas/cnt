#!/dgr/bin/bats -x

@test "Check That we can see the other aci in pod" {
  run /dgr/bin/busybox sh -c "netstat -ltpn | grep ':80'"
  [ "$status" -eq 0 ]
}
