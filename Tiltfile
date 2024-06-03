
# Custom build tool for nix flakes
def build_flake_image(ref, path = "", output = "", resultfile = "result", deps = []):
    build_cmd = "nix build {path}#{output} --refresh --no-link --print-out-paths".format(
        path = path,
        output = output
    )
    commands = [
        "RESULT_IMAGE=$({cmd})".format(cmd = build_cmd),
        "docker image load -i ${RESULT_IMAGE}",
        'IMG_NAME="$(tar -Oxf $RESULT_IMAGE manifest.json | jq -r ".[0].RepoTags[0]")"'.format(ref = ref),
        "docker tag ${IMG_NAME} ${EXPECTED_REF}"
    ]
    custom_build(
        ref,
        command = [
            "nix-shell",
            "--packages",
            "coreutils",
            "gnutar",
            "jq",
            "--run",
            ";\n".join(commands),
        ],
        deps = deps,
    )


image_name = "kurtosistech/kardinal-manager"

cmd_response = local("bash get_host_arch.sh", True, "", True)

# Cleaning up the cmd response
arch_str = str(cmd_response)
arch_str_list = arch_str.splitlines()
host_arch = arch_str_list[0]
print("Host processor architecture: '{}'".format(host_arch))

# Setting the right output depending on the host architecture
output = "containers.aarch64-darwin.kardinal-manager.arm64"
if host_arch == "amd64":
    output = "containers.x86_64-darwin.kardinal-manager.amd64"
print("Using output: {}".format(output))

build_flake_image(image_name , ".", output, deps=["./kontrol-service"])

yaml_dir = "./kontrol-service/deployment"
k8s_yaml(yaml=(yaml_dir + "/k8s.yaml"))

if k8s_context:
    k8s_context(k8s_context)


