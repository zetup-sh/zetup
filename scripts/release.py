import requests
import yaml
import json
import os
from base64 import b64encode

releasesEndpoint = "https://api.github.com/repos/zetup-sh/zetup/releases"

with open("release.yml", "r") as stream:
  try:
    releaseInfo = yaml.safe_load(stream)
  except yaml.YAMLError as exc:
    print(exc)
    exit()

ghUsername = os.environ['ZETUP_GITHUB_USERNAME']
ghToken = os.environ['ZETUP_GITHUB_TOKEN']


def publish_release(releaseInfo):
  response = requests.post(
      url = releasesEndpoint,
      auth=(ghUsername, ghToken),
      data = json.dumps(releaseInfo))
  json_data = json.loads(response.text)
  if json_data.has_key('errors'):
    errors = json_data['errors']
    if errors[0]['code'] == "already_exists":
      if yes_or_no("A release with the tag " + releaseInfo["tag_name"] + " already exists. Would you like to delete it?"):
        delete_release_by_tag(releaseInfo["tag_name"])
        publish_release(releaseInfo)
      else:
        print("not deleting")
        exit()
  else:
    print("release published successfully!")


def yes_or_no(question):
    reply = str(raw_input(question+' (y/n): ')).lower().strip()
    if reply[0] == 'y':
        return True
    if reply[0] == 'n':
        return False
    else:
        return yes_or_no("Uhhhh... please enter ")

def delete_release_by_tag(name):
  releaseId = get_release_id_by_tag(name)
  url = releasesEndpoint+ "/" + str(releaseId)

  response = requests.delete(
      url = url,
      auth=(ghUsername, ghToken))
  if response:
    print("sucessfully deleted old release")
  else:
    print("there was a problem deleting old release")
    exit()

def get_release_id_by_tag(tag):
  url =  releasesEndpoint + "/tags/" + tag
  response = requests.get(
      url = releasesEndpoint,
      auth=(ghUsername, ghToken))
  if not response:
    print("something went wrong", response.text)
    exit()
  else:
    json_data = json.loads(response.text)
    return json_data[0]["id"]

def getSize(filename):
    st = os.stat(filename)
    return st.st_size

addAssetEndpoint = "https://uploads.github.com/repos/zetup-sh/zetup/releases"
def add_assets_to_release(releaseId):
  assets = os.listdir("./build")
  for asset in assets:
    fileLocation = "./build/" + asset
    url = addAssetEndpoint + "/" + str(releaseId) + "/assets?name=" + asset
    print("uploading " + asset)
    with open(fileLocation, 'rb') as f:
      size = getSize(fileLocation)
      headers = {
        'Content-type': 'application/x-binary',
        'Content-length': str(size),
      }

      request = requests.post(
        url,
        auth=(ghUsername, ghToken),
        headers=headers,
        files={'data': f}
      )
      if not request:
        print(request.text)
        exit()
      else:
        print(asset + " uploaded sucessfully")



publish_release(releaseInfo)
releaseId = get_release_id_by_tag(releaseInfo["tag_name"])
add_assets_to_release(releaseId)