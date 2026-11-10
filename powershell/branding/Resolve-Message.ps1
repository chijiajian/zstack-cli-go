$script:EnglishMessages = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'messages.en.json') | ConvertFrom-Json -AsHashtable
$script:ChineseMessages = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'messages.zh-CN.json') | ConvertFrom-Json -AsHashtable

function Resolve-CloudMessage {
    param([string] $Code, [string] $Fallback)

    $messages = if ([System.Globalization.CultureInfo]::CurrentUICulture.TwoLetterISOLanguageName -eq 'zh') { $script:ChineseMessages } else { $script:EnglishMessages }
    $message = if ($messages.Contains($Code)) { $messages[$Code] } else { $Fallback }
    return $messages.prefix -f $script:Package.brand.displayName, $message
}
