export function cleanSemverVersion(version) {
  let shortVersion = version
  if (version.includes("+")) {
    shortVersion = version.split("+")[0]
  }
  return shortVersion
}
