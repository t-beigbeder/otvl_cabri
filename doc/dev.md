# Development useful information

## Support tools

### Coral

[Cobra](https://github.com/spf13/cobra) enables to build CLI for commands and subcommands.

Cobra uses a lot of dependencies useless for this project, we use [coral](https://github.com/muesli/coral). Some
instructions [here](coral.md) for project setup and future reference if needed.

## Audit encrypted repositories workflow

    (odbi *oDssBaseImpl) auditIndex()
      (odbi *oDssBaseImpl) scanStorage()
        (wdi *webDssImpl) scanPhysicalStorage()
          cScanPhysicalStorage()
          (edi *eDssImpl) spScanPhysicalStorageClient()
            (edi *eDssImpl) decryptScannedStorage()
      (odbi *oDssBaseImpl) spLoadRemoteIndex() [empty!]
      (odbi *oDssBaseImpl) doAuditIndexFromStorage()
      (odbi *oDssBaseImpl) doAuditIndexFromIndex()
      (wdi *webDssImpl) spAuditIndexFromRemote()
        (edi *eDssImpl) spLoadRemoteIndex()  
          cLoadIndex()    