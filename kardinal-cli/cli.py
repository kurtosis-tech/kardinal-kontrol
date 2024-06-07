import click
import subprocess


@click.group()
def cli():
    pass


@cli.command()
@click.option("--env", required=True, type=str, help="Environment to deploy to")
@click.argument("service_name")
@click.argument("image_tag")
def create_dev_flow(env, service_name, image_tag):
    flow_id_hash = f"{service_name}-{image_tag}"
    namespace = f"{env}-namespace"

    subprocess.run(["kubectl", "apply", "-k", "prod-mirror", "--namespace", namespace])
    print(f"Deployed with flow ID hash: {flow_id_hash}")


@cli.command()
@click.option("--env", required=True, type=str, help="Environment to delete from")
@click.argument("flow_id_hash")
def delete_dev_flow(env, flow_id_hash):
    namespace = f"{env}-namespace"

    subprocess.run(["kubectl", "delete", "-k", "prod-mirror", "--namespace", namespace])
    print(f"Deleted flow with ID hash: {flow_id_hash}")


if __name__ == "__main__":
    cli()
