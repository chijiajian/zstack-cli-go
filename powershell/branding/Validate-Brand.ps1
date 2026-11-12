function Read-CloudBrand {
    param([Parameter(Mandatory)] [string] $Path)

    $raw = Get-Content -LiteralPath $Path -Raw -ErrorAction Stop
    if (-not (Test-Json -Json $raw -SchemaFile (Join-Path $PSScriptRoot 'brand.schema.json') -ErrorAction Stop)) {
        throw 'Invalid OEM brand configuration.'
    }
    $brand = ConvertFrom-Json -InputObject $raw -AsHashtable -ErrorAction Stop
    if ($brand.displayName -match '[\x00-\x1f\x7f]' -or $brand.publisher -match '[\x00-\x1f\x7f]') { throw 'Brand text must be a single line.' }
    foreach ($key in 'helpUri', 'licenseUri', 'projectUri') {
        if (-not $brand.Contains($key)) { continue }
        $uri = [uri] $brand[$key]
        if (-not $uri.IsAbsoluteUri -or $uri.Scheme -cne 'https' -or -not $uri.Host -or $uri.UserInfo) {
            throw "Brand $key must be an HTTPS URL without user information."
        }
    }
    return $brand
}

function Assert-UniqueCloudBrands {
    param([System.Collections.IDictionary[]] $Brands)

    foreach ($key in 'id', 'moduleName', 'commandPrefix', 'configNamespace', 'moduleGuid') {
        $seen = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
        foreach ($brand in $Brands) {
            if (-not $seen.Add($brand[$key])) { throw "Duplicate OEM brand field: $key" }
        }
    }
}
