# pdns-updater

An updater for PDNS Manager

This program reads the local IPs of the machine it's running and uses the HTTP API of the [PDNS Manager](https://github.com/loewexy/pdnsmanager) to get the external IP of the machine (both IPv4&6) and update a record in the PowerDNS.

## Basic usage

```bash
$ DNSUpdater -ia=1 -iaaaa=2 -ea=3 -eaaaa=4  -pass=password12345 -host=pdnsmanager.instance -i=eth0
```

The program reads the type of the content you want the record to have (internal/external, IPv4/IPv6) from the parameters and updates the one that has the given ID

So for example this

```bash
$ DNSUpdater -ia=1 -pass=password12345 -host=pdnsmanager.instance -i=eth0
```

will update the record that has ID 1 with the IPv4 of the interface eth0.

and this

```bash
$ DNSUpdater -ea=1 -pass=password12345 -host=pdnsmanager.instance -i=eth0
```

will update the record that has ID 1 with the external IPv4 of the machine.

You can also combine different parameters

```bash
$ DNSUpdater -ia=1 -eaaaa=2 -pass=password12345 -host=pdnsmanager.instance -i=eth0
```

This will set the record with ID 1 the internal IPv4 of the eth0 and the record with ID 2 the external IPv6.

### Parameters list

|       |                                                                                                                    |
| ----- | ------------------------------------------------------------------------------------------------------------------ |
| ia    | The ID of the internal A record you want to update                                                                 |
| iaaaa | The ID of the internal AAAA record you want to update                                                              |
| ea    | The ID of the external A record you want to update                                                                 |
| eaaaa | The ID of the external AAAA record you want to update                                                              |
| pass  | The password used to authenticate. Can be created in the PDNS Manager by clicking the key icon on the domains list |
| host  | The URL that the PDNS instance can be found                                                                        |
| i     | The interface to get the internal IPs                                                                              |

## Usage with cron

In this repo you can also find a file called `update.sh`. It's a simple wrapper with the parameters extracted as variables. You can place that and the executable of the program in the same folder and update your Dynamic DNS records easier with cron.

For example I use it like this

```bash
*/1 * * * * ~/.crons/ddns/update.sh
```

and every minute the script updates the IPs of my machine

## Compile

To compile this program you need the [Go compiler](https://golang.org/). Then, run `go build` inside the folder that contains the `main.go` file.
