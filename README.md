# KVBench

cloned from [tidwall/kvbench](https://github.com/tidwall/kvbench)

KVBench is a Redis server clone backed by a few different Go databases. 

It's intended to be used with the `redis-benchmark` command to test the
performance of various Go databases.  It has support for redis pipelining. The
`redis-benchmark` can run as explained here https://github.com/tidwall/kvbench#examples.

This cloned version adds more kv databases and automatic scripts.

Features:

- Databases
  - [badger](https://github.com/dgraph-io/badger)
  - [BboltDB](https://go.etcd.io/bbolt)
  - [BoltDB](https://github.com/boltdb/bolt)
  - [buntdb](https://github.com/tidwall/buntdb)
  - [LevelDB](https://github.com/syndtr/goleveldb)
  - [cznic/kv](https://github.com/cznic/kv)
  - [rocksdb](https://github.com/tecbot/gorocksdb)
  - [pebble](https://github.com/cockroachdb/pebble)
  - [pogreb](https://github.com/akrylysov/pogreb)
  - [nutsdb](https://github.com/xujiajun/nutsdb)
  - [sniper](https://github.com/recoilme/sniper)
  - map (in-memory) with [AOF persistence](https://redis.io/topics/persistence)
  - btree (in-memory) with [AOF persistence](https://redis.io/topics/persistence)
- Option to disable fsync
- Compatible with Redis clients


## SSD benchmark
The following benchmarks show the throughput of inserting/reading keys (of size
9 bytes) and values (of size 256 bytes).

### nofsync

**throughputs**

| |badger|bbolt|bolt|leveldb|kv|buntdb|pebble|pogreb|nutsdb|rocksdb|btree|btree/memory|map|map/memory|
|--|--|--|--|--|--|--|--|--|--|--|--|--|--|--|
|Batch Write|1818000|305000|279000|458000|113000|1109000|484000|1694000|1978000|440000|2322000|OOM|2718000|OOM|
|del|42068|5830|5602|72084|124984|47913|92244|4264|114012|86168|597002|761315|1289271|3956378|
|set|39945|17457|17584|119846|11598|40720|77794|66457|113455|81718|137413|652104|137153|551238|
|get|474943|691812|706825|416832|32969|1857733|321527|5063794|630274|295142|3034517|2035675|7752284|6459314|
|setmixed|12353|10525|9520|21094|5133|9590|58853|50422|28863|57994|55785|86495|85366|134747|
|getmixed|171771|448769|469127|183170|20539|49857|168025|210983|137648|172137|232248|366470|377661|697788|

**time (latency)**

| |badger|bbolt|bolt|leveldb|kv|buntdb|pebble|pogreb|nutsdb|rocksdb|btree|btree/memory|map|map/memory|
|--|--|--|--|--|--|--|--|--|--|--|--|--|--|--|
|getmixed|1455|557|532|1364|12171|5014|1487|1184|1816|1452|1076|682|661|358|
|set|6258|14320|14216|2086|21554|6139|3213|3761|2203|3059|1819|383|1822|453|
|get|526|361|353|599|7582|134|777|49|396|847|82|122|32|38|
|setmixed|80945|95010|105037|47405|194793|104264|16991|19832|34646|17243|17925|11561|11714|7421|
|del|5942|42880|44625|3468|2000|5217|2710|58624|2192|2901|418|328|193|63|

### fsync

**throughputs**

| |badger|bbolt|bolt|leveldb|kv|buntdb|pebble|pogreb|nutsdb|rocksdb|btree|btree/memory|map|map/memory|
|--|--|--|--|--|--|--|--|--|--|--|--|--|--|--|
|Batch Write|1981000|332000|308000|122000|206000|759000|324000|16000|26000|392000|1350000|OOM|1510000|OOM|
|get|511262|740201|749409|1399345|33044|2419811|312589|7256537|1762448|279579|4073015|1931155|12368046|5064254|
|setmixed|9345|52|51|83|5151|97|2648|49|97|2809|94|74675|93|124593|
|getmixed|41121|712940|714763|725069|20606|1756|274968|557|435|272844|1002|318337|1963|638967|
|del|19595|49|49|84619|12190|97|29386|961261|98|32109|49337|805323|94|3419262|
|set|18278|48|48|236|12850|97|29000|52|97|28579|93|535504|94|467439|


**time (latency)**

| |badger|bbolt|bolt|leveldb|kv|buntdb|pebble|pogreb|nutsdb|rocksdb|btree|btree/memory|map|map/memory|
|--|--|--|--|--|--|--|--|--|--|--|--|--|--|--|
|setmixed|106999|19187738|19329929|12000867|194129|10261633|377610|20268643|10255662|355960|10620359|13391|10652140|8026|
|get|488|337|333|178|7565|103|799|34|141|894|61|129|20|49|
|getmixed|6079|350|349|344|12132|142337|909|448760|574240|916|249360|785|127325|391|
|del|12758|5078760|5095016|2954|20508|2577078|8507|260|2550875|7785|5067|310|2654177|73|
|set|13676|5152682|5167592|1056019|19454|2571839|8620|4736548|2558664|8747|2683904|466|2649372|534|

