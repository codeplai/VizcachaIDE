"""VizcachaIDE fix for Rust enums in LLDB on *-gnu targets (docs/PLAN_RUST.md, M3 QA).

LLDB 23 turns rustc's DWARF variant parts into a "$variants$" union whose variants hold
"$discr$" and "value", but it places "value" after the variant struct instead of at offset 0, so
Some(7) showed garbage (Some(440)). rustc's payload structs carry the absolute offsets of their
fields from the start of the enum (Some<i32>.__0 is at 4), so reading the payload at the enum's own
address gives the right fields. Imported after rustc's lldb_lookup.py.
"""
import lldb
import lldb_providers

_original_update = lldb_providers.ClangEncodedEnumProvider.update


def _update(self):
    _original_update(self)
    payload = self.value.GetNonSyntheticValue() if self.value.IsValid() else self.value
    address = self.valobj.GetLoadAddress()
    if not payload.IsValid() or address == lldb.LLDB_INVALID_ADDRESS:
        return
    fixed = self.valobj.CreateValueFromAddress("value", address, payload.GetType())
    if not fixed.IsValid():
        return
    synthetic = fixed.GetSyntheticValue()
    self.value = synthetic if synthetic.IsValid() else fixed


lldb_providers.ClangEncodedEnumProvider.update = _update


def __lldb_init_module(debugger, _dict):
    pass
