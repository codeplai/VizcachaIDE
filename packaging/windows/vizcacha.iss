; Inno Setup 6.3+ script for VizcachaIDE.
; Normally compiled by `python packaging/build.py --variant full|lite --package`, which passes:
;   ISCC /DAppVersion=0.2.0 /DAppVersionNumeric=0.2.0.0 /DSourceDir=<dist\...\VizcachaIDE>
;        /DVariant=full /DLicenseFile=<LICENSE-NOTICE.txt> /DOutputDir=<dist\release>
;        /DOutputBaseName=<name> packaging\windows\vizcacha.iss
;
; - Per-user install by default (no admin); the first page lets the user choose
;   "Install for all users" (PrivilegesRequiredOverridesAllowed=dialog).
; - Setup languages: English and Spanish (auto-detected from the Windows UI language).
; - License page: GPLv3 distribution notice (packaging/NOTICE.md) + the MIT license.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef AppVersionNumeric
  #define AppVersionNumeric "0.0.0.0"
#endif
#ifndef SourceDir
  #define SourceDir "..\..\dist\lite\VizcachaIDE"
#endif
#ifndef Variant
  #define Variant "lite"
#endif
#ifndef LicenseFile
  #define LicenseFile "..\..\build\installer\LICENSE-NOTICE.txt"
#endif
#ifndef OutputDir
  #define OutputDir "..\..\dist\release"
#endif
#ifndef OutputBaseName
  #define OutputBaseName "VizcachaIDE-" + AppVersion + "-windows-x64-" + Variant + "-setup"
#endif

#define AppName "VizcachaIDE"
#define AppExe "VizcachaIDE.exe"
#define ProgId "VizcachaIDE.go"

[Setup]
AppId={{4BEE5965-6B1F-4655-B971-49D37B2FC164}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion} ({#Variant})
AppPublisher=Codeplai Games
AppCopyright=Copyright (c) 2025-2026 Marks Calderon - Codeplai Games
VersionInfoVersion={#AppVersionNumeric}
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog commandline
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
LicenseFile={#LicenseFile}
OutputDir={#OutputDir}
OutputBaseFilename={#OutputBaseName}
SetupIconFile={#SourceDir}\_internal\logo.ico
UninstallDisplayIcon={app}\{#AppExe}
UninstallDisplayName={#AppName} {#AppVersion}
Compression=lzma2/max
SolidCompression=yes
WizardStyle=modern
ChangesAssociations=yes
ShowLanguageDialog=auto
LanguageDetectionMethod=uilanguage

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "spanish"; MessagesFile: "compiler:Languages\Spanish.isl"

[CustomMessages]
english.AssocGroup=File associations:
spanish.AssocGroup=Asociaciones de archivos:
english.AssociateGo=Open .go files with VizcachaIDE
spanish.AssociateGo=Abrir los archivos .go con VizcachaIDE
english.GoSourceFile=Go source file
spanish.GoSourceFile=Archivo de código fuente Go

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "associatego"; Description: "{cm:AssociateGo}"; GroupDescription: "{cm:AssocGroup}"; Flags: unchecked

[InstallDelete]
; Switching from "full" to "lite" (same AppId) must not leave a stale bundled toolchain behind.
Type: filesandordirs; Name: "{app}\toolchain"
Type: filesandordirs; Name: "{app}\_internal"

[Files]
Source: "{#SourceDir}\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\{#AppExe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; Tasks: desktopicon

[Registry]
; HKA = HKCU for per-user installs, HKLM for all-users installs.
Root: HKA; Subkey: "Software\Classes\.go\OpenWithProgids"; ValueType: string; ValueName: "{#ProgId}"; ValueData: ""; Flags: uninsdeletevalue; Tasks: associatego
Root: HKA; Subkey: "Software\Classes\{#ProgId}"; ValueType: string; ValueName: ""; ValueData: "{cm:GoSourceFile}"; Flags: uninsdeletekey; Tasks: associatego
Root: HKA; Subkey: "Software\Classes\{#ProgId}\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#AppExe},0"; Tasks: associatego
Root: HKA; Subkey: "Software\Classes\{#ProgId}\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#AppExe}"" ""%1"""; Tasks: associatego
Root: HKA; Subkey: "Software\Classes\Applications\{#AppExe}\SupportedTypes"; ValueType: string; ValueName: ".go"; ValueData: ""; Flags: uninsdeletekey; Tasks: associatego

[Run]
Filename: "{app}\{#AppExe}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent
