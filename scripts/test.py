import sys
import random
import requests
import string
import time

uris = [
    ''.join(
        random.choice(string.ascii_letters)
        for _ in range(random.randint(5, 15))
    ) for _ in range(100)
]

assets = []

API_BASE = '{}/api/v1'.format(sys.argv[1])

print('Checking that we can create assets...')
for i, uri in enumerate(uris):
    resp = requests.post('{}/assets'.format(API_BASE), json={
        'data': {
            'type': 'assets',
            'attributes': {
                'uri': 'uri://{}'.format(uri),
                'name': 'URI #{} ({})'.format(i, uri),
            }
        }
    })

    assert resp.status_code == 201
    assets.append(resp.json()['data'])
print('Looks good\n')

print('Checking that we can\'t create duplicates...')
for i, uri in enumerate(uris):
    resp = requests.post('{}/assets'.format(API_BASE), json={
        'data': {
            'type': 'assets',
            'attributes': {
                'uri': 'uri://{}'.format(uri),
                'name': 'URI #{} ({})'.format(i, uri),
            }
        }
    })

    assert resp.status_code == 409
    assert resp.json() ==  {
        'errors': [
            {'status': '409', 'detail': 'UNIQUE constraint failed: asset.id', 'title': 'Conflict'}
        ]
    }
print('Looks good\n')

print('Checking that we can add notes...')
for i, uri in enumerate(uris):
    resp = requests.post('{}/notes'.format(API_BASE), json={
        'data': {
            'type': 'notes',
            'attributes': {
                'content': 'THIS IS MY NOTE'
            },
            'relationships': {
                'asset': {
                    'data': {'type': 'assets', 'id': assets[i]['id']}
                }
            }
        }
    })
    assert resp.status_code == 201
print('Looks good\n')


print('Checking that listing records matches what we\'ve made...')
resp = requests.get('{}/assets'.format(API_BASE))
assert resp.status_code == 200
assert resp.json()['data'] == assets
print('Looks good\n')


print('Checking that we can get specific assets...')
for asset in assets:
    resp = requests.get('{}/assets/{}'.format(API_BASE, asset['id']))
    assert resp.status_code == 200
    assert resp.json()['data'] == asset
print('Looks good\n')

if len(sys.argv) > 2:
    API_BASE_2 = '{}/api/v1'.format(sys.argv[2])

    print('Checking that replication to {} works...'.format(API_BASE_2))
    print('Waiting 10 seconds...')
    time.sleep(10)
    for asset in assets:
        resp = requests.get('{}/assets/{}'.format(API_BASE_2, asset['id']))
        assert resp.status_code == 200
        assert resp.json()['data'] == asset
    print('Looks good\n')


print('Checking that we can delete assets...')
for asset in assets:
    resp = requests.delete('{}/assets/{}'.format(API_BASE, asset['id']))
    assert resp.status_code == 204
print('Looks good\n')

print('Checking that we can\'t double delete assets...')
for asset in assets:
    resp = requests.get('{}/assets/{}'.format(API_BASE, asset['id']))
    assert resp.status_code == 410
print('Looks good\n')

print('Checking that we can\'t add notes to deleted assets...')
for asset in assets:
    resp = requests.post('{}/notes'.format(API_BASE), json={
        'data': {
            'type': 'notes',
            'attributes': {
                'content': 'THIS IS MY NOTE'
            },
            'relationships': {
                'asset': {
                    'data': {'type': 'assets', 'id': asset['id']}
                }
            }
        }
    })
    assert resp.status_code == 406
print('Looks good\n')

print('Checking that everything has been deleted...')
resp = requests.get('{}/assets'.format(API_BASE))
assert len(resp.json()['data']) == 0
print('Looks good\n')
