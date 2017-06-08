import random
import requests
import string

uris = [
    ''.join(
        random.choice(string.ascii_letters)
        for _ in range(random.randint(5, 15))
    ) for _ in range(100)
]


for i, uri in enumerate(uris):
    resp = requests.post('http://localhost:8080/api/v1/assets', json={
        'data': {
            'type': 'assets',
            'attributes': {
                'uri': 'uri://{}'.format(uri),
                'name': 'URI #{} ({})'.format(i, uri),
            }
        }
    })
    print(resp)
    try:
        print(resp.json())
    except Exception:
        pass

    resp = requests.post('http://localhost:8080/api/v1/notes', json={
        'data': {
            'type': 'notes',
            'attributes': {
                'content': 'THIS IS MY NOTE FOR {}'.format(i)
            },
            'relationships': {
                'asset': {
                    'data': {'type': 'assets', 'id': resp.json()['data']['id']}
                }
            }
        }
    })
    assert resp.status_code == 201

    print(resp)
    try:
        print(resp.json())
    except Exception:
        pass
