# gridusagereindex

Command line tool to reindex Usage for the [Grid](https://github.com/guardian/grid) service from the [Guardian Content Platform](http://open-platform.theguardian.com/).

---

## Usage Example

Requires `config.json` in the same folder as the executable:

```json
{
  "CapiUrl": "https://content.guardianapis.com/",
  "CapiApiKey": "my-capi-api-key",
  "UsageUrl": "https://usage.example.com",
  "GridApiKey": "my-grid-api-key",
  "GobbyFile": "/var/tmp/gobbyfile",
  "FromDate": "2014-02-16",
  "ToDate": "2014-02-17"
}
```
