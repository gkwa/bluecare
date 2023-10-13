import json
import pathlib

import requests
import yaml

url = "https://raw.githubusercontent.com/aws/aws-sdk-net/master/sdk/src/Core/endpoints.json"

local_json_path = "endpoints.json"
local_yaml_path = "endpoints.yaml"

if not pathlib.Path(local_json_path).exists():
    response = requests.get(url)

    if response.status_code == 200:
        with open(local_json_path, "wb") as file:
            file.write(response.content)
    else:
        print("Failed to download the JSON file.")
        exit()

with open(local_json_path, "r") as json_file:
    data = json.load(json_file)

new_service_names = list(data["partitions"][0]["services"].keys())

if pathlib.Path(local_yaml_path).exists():
    with open(local_yaml_path, "r") as yaml_file:
        existing_data = yaml.safe_load(yaml_file)
else:
    existing_data = {}

if "service_names" not in existing_data:
    existing_data["service_names"] = []

for new_service in new_service_names:
    if new_service not in existing_data["service_names"]:
        existing_data["service_names"].append(new_service)

with open(local_yaml_path, "w") as yaml_file:
    yaml.dump(existing_data, yaml_file, default_flow_style=False)

print(f"New service names added: {new_service_names}")

print("Merged Data:")
print(yaml.dump(existing_data, default_flow_style=False))
