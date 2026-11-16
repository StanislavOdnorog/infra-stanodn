"""Update only the existing n8n A record; credentials remain in GitHub secrets."""
import json
import os
import pathlib
import urllib.request

record = next(r for r in json.loads(pathlib.Path('terraform/cloudflare/stanodn.org/dns-records.json').read_text()) if r['name'] == 'n8n' and r['type'] == 'A')
base = 'https://api.cloudflare.com/client/v4/zones/' + os.environ['CF_ZONE_ID'] + '/dns_records'
headers = {'Authorization': 'Bearer ' + os.environ['CF_API_TOKEN'], 'Content-Type': 'application/json'}

def call(url, data=None):
    req = urllib.request.Request(url, headers=headers, data=None if data is None else json.dumps(data).encode(), method='GET' if data is None else 'PATCH')
    result = json.load(urllib.request.urlopen(req))
    assert result['success'], 'Cloudflare request failed'
    return result['result']

records = call(base + '?type=A&name=n8n.stanodn.org')
assert len(records) == 1, 'Expected exactly one existing n8n A record'
desired = dict(type='A', name='n8n.stanodn.org', content=record['value'], proxied=record['proxied'], ttl=record['ttl'], comment=record['comment'])
if any(records[0].get(k) != v for k, v in desired.items()):
    call(base + '/' + records[0]['id'], desired)
live = call(base + '/' + records[0]['id'])
assert all(live.get(k) == v for k, v in desired.items())
print('Verified n8n.stanodn.org ->', live['content'], 'proxied:', live['proxied'])
