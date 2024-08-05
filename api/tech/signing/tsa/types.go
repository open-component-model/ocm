package tsa

import (
	cms "github.com/InfiniteLoopSpace/go_S-MIME/cms/protocol"
	tsa "github.com/InfiniteLoopSpace/go_S-MIME/timestamp"
)

type TimeStamp = cms.SignedData

type MessageImprint = tsa.MessageImprint
