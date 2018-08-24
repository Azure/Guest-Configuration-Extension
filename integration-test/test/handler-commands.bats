#!/usr/bin/env bats

load test_helper

setup(){
    build_docker_image
}

teardown(){
    rm -rf "$certs_dir"
}

@test "handler command: enable - can enable the gc agent" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable"
    push_settings '' ''

    run start_container

     # Validate .status file says enable failed
     diff="$(container_diff)"; echo "$diff"
    [[ "$diff" = *"A /var/lib/waagent/Extension/status/0.status"* ]]
    status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
    echo "$status_file"; [[ "$status_file" = *'enable succeeded'* ]]
}

@test "handler command: disable - can disable the agent" {
    mk_container sh -c "fake-waagent install && fake-waagent enable && wait-for-enable && fake-waagent disable && wait-for-disable"
    push_settings '' ''

    run start_container

    # Validate .status file says disable succeeded
     diff="$(container_diff)"; echo "$diff"
     [[ "$diff" = *"A /var/lib/waagent/Extension/status/0.status"* ]]
     status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
     echo "$status_file"; [[ "$status_file" = *'disable succeeded'* ]]
}

#@test "handler command: update - can update the gc agent" {
#    mk_container sh -c "fake-waagent update && wait-for-update"
#    push_settings '' ''
#
#   run start_container
#
#    # Validate .status file says update succeeded
#     diff="$(container_diff)"; echo "$diff"
#     [[ "$diff" = *"A /var/lib/waagent/Extension/status/0.status"* ]]
#     status_file="$(container_read_file /var/lib/waagent/Extension/status/0.status)"
#     echo "$status_file"; [[ "$status_file" = *'update succeeded'* ]]
#}

# @test "handler command: uninstall - deletes the data dir" {
#     run in_container sh -c \
#         "fake-waagent install && fake-waagent enable && fake-waagent uninstall"
#     echo "$output"
#     [ "$status" -eq 0 ]
#
#     diff="$(container_diff)" && echo "$diff"
#     [[ "$diff" != */var/lib/waagent/guest-configuration* ]]
# }
