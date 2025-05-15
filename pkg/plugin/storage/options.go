package storage

import (
	"time"
)

type IBeforeInitiateMultipartUpload interface {
	BeforeInitiateMultipartUpload(originalName, group string) error
}

type IAfterInitiateMultipartUpload interface {
	AfterInitiateMultipartUpload(result *InitiateMultipartUploadResult) error
}

type IBeforeCompleteMultipartUpload interface {
	BeforeCompleteMultipartUpload(uploadID, objectKey string, parts []UploadPart) error
}

type IAfterCompleteMultipartUpload interface {
	AfterCompleteMultipartUpload(result *CompleteMultipartUploadResult) error
}

type IBeforeGeneratePublicURL interface {
	BeforeGeneratePublicURL(objectKey string) error
}

type IAfterGeneratePublicURL interface {
	AfterGeneratePublicURL(url string) error
}

type IBeforeGenerateUploadPartURL interface {
	BeforeGenerateUploadPartURL(uploadID, objectKey string, partNumber int, expires time.Duration) error
}

type IAfterGenerateUploadPartURL interface {
	AfterGenerateUploadPartURL(result *UploadPartInfo) error
}

type IBeforeDeleteObject interface {
	BeforeDeleteObject(objectKey string) error
}

type IAfterDeleteObject interface {
	AfterDeleteObject(objectKey string) error
}

type IFileManagerHook interface {
	IBeforeInitiateMultipartUpload
	IBeforeCompleteMultipartUpload
	IBeforeGeneratePublicURL
	IBeforeGenerateUploadPartURL
	IBeforeDeleteObject
	IAfterInitiateMultipartUpload
	IAfterCompleteMultipartUpload
	IAfterGeneratePublicURL
	IAfterGenerateUploadPartURL
	IAfterDeleteObject
}

func WithFileManagerHook(f IFileManagerHook) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeInitiateMultipartUpload = f.BeforeInitiateMultipartUpload
		fm.BeforeCompleteMultipartUpload = f.BeforeCompleteMultipartUpload
		fm.BeforeGeneratePublicURL = f.BeforeGeneratePublicURL
		fm.BeforeGenerateUploadPartURL = f.BeforeGenerateUploadPartURL
		fm.BeforeDeleteObject = f.BeforeDeleteObject
		fm.AfterInitiateMultipartUpload = f.AfterInitiateMultipartUpload
		fm.AfterCompleteMultipartUpload = f.AfterCompleteMultipartUpload
		fm.AfterGeneratePublicURL = f.AfterGeneratePublicURL
		fm.AfterGenerateUploadPartURL = f.AfterGenerateUploadPartURL
		fm.AfterDeleteObject = f.AfterDeleteObject
	}
}

func WithBeforeInitiateMultipartUpload(f IBeforeInitiateMultipartUpload) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeInitiateMultipartUpload = f.BeforeInitiateMultipartUpload
	}
}

func WithBeforeInitiateMultipartUploadFun(f func(originalName, group string) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeInitiateMultipartUpload = f
	}
}

func WithAfterInitiateMultipartUpload(f IAfterInitiateMultipartUpload) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterInitiateMultipartUpload = f.AfterInitiateMultipartUpload
	}
}

func WithAfterInitiateMultipartUploadFun(f func(result *InitiateMultipartUploadResult) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterInitiateMultipartUpload = f
	}
}

func WithBeforeCompleteMultipartUpload(f IBeforeCompleteMultipartUpload) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeCompleteMultipartUpload = f.BeforeCompleteMultipartUpload
	}
}

func WithBeforeCompleteMultipartUploadFun(f func(uploadID, objectKey string, parts []UploadPart) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeCompleteMultipartUpload = f
	}
}

func WithAfterCompleteMultipartUpload(f IAfterCompleteMultipartUpload) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterCompleteMultipartUpload = f.AfterCompleteMultipartUpload
	}
}

func WithAfterCompleteMultipartUploadFun(f func(result *CompleteMultipartUploadResult) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterCompleteMultipartUpload = f
	}
}

func WithBeforeGeneratePublicURL(f IBeforeGeneratePublicURL) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeGeneratePublicURL = f.BeforeGeneratePublicURL
	}
}

func WithBeforeGeneratePublicURLFun(f func(objectKey string) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeGeneratePublicURL = f
	}
}

func WithAfterGeneratePublicURL(f IAfterGeneratePublicURL) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterGeneratePublicURL = f.AfterGeneratePublicURL
	}
}

func WithAfterGeneratePublicURLFun(f func(url string) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterGeneratePublicURL = f
	}
}

func WithBeforeGenerateUploadPartURL(f IBeforeGenerateUploadPartURL) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeGenerateUploadPartURL = f.BeforeGenerateUploadPartURL
	}
}

func WithBeforeGenerateUploadPartURLFun(f func(uploadID, objectKey string, partNumber int, expires time.Duration) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeGenerateUploadPartURL = f
	}
}

func WithAfterGenerateUploadPartURL(f IAfterGenerateUploadPartURL) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterGenerateUploadPartURL = f.AfterGenerateUploadPartURL
	}
}

func WithAfterGenerateUploadPartURLFun(f func(result *UploadPartInfo) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterGenerateUploadPartURL = f
	}
}

func WithBeforeDeleteObject(f IBeforeDeleteObject) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeDeleteObject = f.BeforeDeleteObject
	}
}

func WithBeforeDeleteObjectFun(f func(objectKey string) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.BeforeDeleteObject = f
	}
}

func WithAfterDeleteObject(f IAfterDeleteObject) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterDeleteObject = f.AfterDeleteObject
	}
}

func WithAfterDeleteObjectFun(f func(objectKey string) error) FileManagerHookOption {
	return func(fm *fileManagerWithHook) {
		fm.AfterDeleteObject = f
	}
}
