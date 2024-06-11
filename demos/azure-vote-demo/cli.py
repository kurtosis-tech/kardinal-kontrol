#! /usr/bin/env nix-shell
#! nix-shell -i python3 -p python3 python3Packages.click

import click
import subprocess
import os

file_dir = os.path.dirname(os.path.abspath(__file__))


@click.group()
def cli():
    pass


def replace_pod(namespace):
    try:
        get_cmd = [
            "kubectl",
            "get",
            "pod",
            "-n",
            namespace,
            "-l",
            "app=redis-proxy-overlay",
            "-o",
            "yaml",
        ]

        replace_cmd = ["kubectl", "replace", "--force", "-n", namespace, "-f", "-"]

        get_proc = subprocess.Popen(get_cmd, stdout=subprocess.PIPE)
        replace_proc = subprocess.Popen(
            replace_cmd,
            stdin=get_proc.stdout,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        get_proc.stdout.close()
        out, err = replace_proc.communicate()

        if replace_proc.returncode != 0:
            print(f"Error: {err.decode()}")
        else:
            print(f"Output: {out.decode()}")

    except subprocess.CalledProcessError as e:
        print(f"Command failed with return code {e.returncode}")


def delete_dev_resources(resource, namespace):
    try:
        get_cmd = [
            "kubectl",
            "get",
            resource,
            "-n",
            namespace,
            "-l",
            "type=dev",
            "-o",
            "yaml",
        ]

        delete_cmd = ["kubectl", "delete", "-n", namespace, "-f", "-"]

        get_proc = subprocess.Popen(get_cmd, stdout=subprocess.PIPE)
        delete_proc = subprocess.Popen(
            delete_cmd,
            stdin=get_proc.stdout,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        get_proc.stdout.close()
        out, err = delete_proc.communicate()

        if delete_proc.returncode != 0:
            print(f"Error: {err.decode()}")
        else:
            print(f"Output: {out.decode()}")

    except subprocess.CalledProcessError as e:
        print(f"Command failed with return code {e.returncode}")


@cli.command()
@click.option("--env", required=True, type=str, help="Environment to deploy to")
@click.argument("namespace")
@click.argument("image_tag")
def create_dev_flow(env, namespace, image_tag):
    flow_id_hash = f"{namespace}"

    subprocess.run(
        [
            "kubectl",
            "apply",
            "-f",
            f"{file_dir}/dev-in-prod-demo.yaml",
            "--namespace",
            namespace,
        ]
    )
    print(f"Deployed with flow ID hash: {flow_id_hash}")


@cli.command()
@click.option("--env", required=True, type=str, help="Environment to delete from")
@click.argument("flow_id_hash")
def delete_dev_flow(env, flow_id_hash):
    namespace = f"{flow_id_hash}"

    subprocess.run(
        ["kubectl", "apply", "-n", namespace, "-f", f"{file_dir}/prod-only-demo.yaml"]
    )

    for command in ["all", "virtualservices", "destinationrules"]:
        delete_dev_resources(command, namespace)

    print(f"Deleted flow with ID hash: {flow_id_hash}")


@cli.command()
@click.option("--env", required=True, type=str, help="Environment to deploy to")
@click.argument("flow_id_hash")
def reset_dev_flow(env, flow_id_hash):
    namespace = f"{flow_id_hash}"
    replace_pod(namespace)


if __name__ == "__main__":
    cli()
