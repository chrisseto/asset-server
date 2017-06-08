import json
import random
import string
import base64


uris = [
    ''.join(
        random.choice(string.ascii_letters)
        for _ in range(random.randint(5, 15))
    ) for _ in range(100)
]

print('GET http://localhost:8080/api/v1/assets')
print('GET http://localhost:8080/api/v1/notes')
print('PUT http://localhost:8080/api/v1/notes')
print('PUT http://localhost:8080/api/v1/assets')

for i, uri in enumerate(uris):
    asset = {
            'data': {
                'type': 'assets',
                'attributes': {
                    'uri': 'uri://{}'.format(uri),
                    'name': 'URI #{} ({})'.format(i, uri),
                }
            }
        }

    with open('./scripts/bodies/{:04}.json'.format(i), 'w') as fobj:
        json.dump(asset, fobj)

    print('POST http://localhost:8080/api/v1/assets')
    print('@./scripts/bodies/{:04}.json\n'.format(i))
    print('GET http://localhost:8080/api/v1/assets/{}'.format(base64.b64encode(asset['data']['attributes']['uri']).strip('=')))
    print('DELETE http://localhost:8080/api/v1/assets/{}'.format(base64.b64encode(asset['data']['attributes']['uri']).strip('=')))
