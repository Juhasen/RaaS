<#
One-off: rewrites media URLs stored as http://localhost:9000/raas-media-bucket/<key> to the
same-origin /media-files/<key> that the gateway serves. Safe to re-run (matches only old URLs).
Uses the MongoDB URL the services already use (ConfigMap raas-config).
#>
param ([string]$ConfigMap = 'raas-config')

$ErrorActionPreference = 'Continue'

$uri = kubectl -n raas get configmap $ConfigMap -o 'jsonpath={.data.MONGODB_URL}'
if ($LASTEXITCODE -ne 0 -or -not $uri) { throw "MONGODB_URL not found in ConfigMap $ConfigMap." }

$js = @'
const old = "http://localhost:9000/raas-media-bucket/";
const fresh = "/media-files/";
for (const dbName of db.getMongo().getDBNames().filter(n => !["admin", "config", "local"].includes(n))) {
  const d = db.getSiblingDB(dbName);
  for (const c of d.getCollectionNames()) {
    const coll = d.getCollection(c);
    if (coll.countDocuments({ media_urls: { $regex: "^http://localhost:9000/" } }) > 0) {
      let n = 0;
      coll.find({ media_urls: { $regex: "^http://localhost:9000/" } }).forEach(doc => {
        coll.updateOne({ _id: doc._id }, { $set: { media_urls: doc.media_urls.map(u => u.startsWith(old) ? fresh + u.slice(old.length) : u) } });
        n++;
      });
      print(dbName + "." + c + " media_urls updated: " + n);
    }
    if (coll.countDocuments({ url: { $regex: "^http://localhost:9000/" } }) > 0) {
      const r = coll.updateMany({ url: { $regex: "^http://localhost:9000/raas-media-bucket/" } },
        [{ $set: { url: { $concat: [fresh, { $substrCP: ["$url", old.length, 2000] }] } } }]);
      print(dbName + "." + c + " url updated: " + r.modifiedCount);
    }
  }
}
'@

$js | kubectl -n raas exec -i deploy/mongodb -- mongosh $uri --quiet
if ($LASTEXITCODE -ne 0) { throw 'mongosh failed.' }
