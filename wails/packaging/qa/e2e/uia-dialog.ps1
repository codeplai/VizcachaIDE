# Answers VizcachaIDE's native file dialog ("Abrir archivo", "Guardar archivo") with UI Automation:
# finds the dialog by its title, writes the file name and presses its OK button.
# No keystrokes and no mouse: nothing else on the desktop is touched.
#   powershell -File uia-dialog.ps1 -Title "Abrir archivo" -FilePath C:\x\a.go
param(
    [Parameter(Mandatory)] [string] $Title,
    [Parameter(Mandatory)] [string] $FilePath,
    [int] $Seconds = 30
)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName UIAutomationClient, UIAutomationTypes
$A = [System.Windows.Automation.AutomationElement]
$T = [System.Windows.Automation.TreeScope]
$C = [System.Windows.Automation.PropertyCondition]

$byTitle = New-Object $C ($A::NameProperty, $Title)
$dialog = $null
$deadline = (Get-Date).AddSeconds($Seconds)
while (-not $dialog -and (Get-Date) -lt $deadline) {
    # The dialog is a top-level window, or a child of the app window that owns it.
    $dialog = $A::RootElement.FindFirst($T::Children, $byTitle)
    if (-not $dialog) {
        foreach ($w in $A::RootElement.FindAll($T::Children, [System.Windows.Automation.Condition]::TrueCondition)) {
            $owned = $w.FindFirst($T::Children, $byTitle)
            if ($owned) { $dialog = $owned; break }
        }
    }
    if (-not $dialog) { Start-Sleep -Milliseconds 300 }
}
if (-not $dialog) { throw "dialog '$Title' did not open" }

# 1148 is the file name box of the common file dialog; 1 is its OK button (Abrir / Guardar).
$edit = $dialog.FindFirst($T::Descendants, (New-Object $C ($A::AutomationIdProperty, '1148')))
if (-not $edit) { throw 'file name box not found' }
$edit.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern).SetValue($FilePath)
Start-Sleep -Milliseconds 300
$ok = $dialog.FindFirst($T::Children, (New-Object $C ($A::AutomationIdProperty, '1')))
$ok.GetCurrentPattern([System.Windows.Automation.InvokePattern]::Pattern).Invoke()
"answered '$Title' with $FilePath"
