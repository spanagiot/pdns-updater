host=pdnsmanager.instance
password=password123456
interface=eth0
InternalAID=1
InternalAAAAID=
ExternalAID=
ExternalAAAAID=


./DNSUpdater -ia="$InternalAID" -iaaaa="$InternalAAAAID" -ea="$ExternalAID" -eaaaa="$ExternalAAAAID" -i="$interface" --pass="$password" --host="$host"
