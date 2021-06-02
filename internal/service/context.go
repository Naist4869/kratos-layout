package service

import (
	"context"
	"net/http"
	"reflect"
)

type serviceKey struct{}

type ServiceInfo struct {
	RedirectFunc func(w http.ResponseWriter, r *http.Request)
}

func ContextServiceKey(ctx context.Context) *ServiceInfo {
	biz, _ := ctx.Value(serviceKey{}).(*ServiceInfo)
	return biz
}

func WithServiceKey(ctx context.Context, info *ServiceInfo) context.Context {
	if info == nil {
		panic("nil info")
	}
	old := ContextServiceKey(ctx)
	info.compose(old)
	ctx = context.WithValue(ctx, serviceKey{}, info)
	return ctx
}

func (b *ServiceInfo) compose(old *ServiceInfo) {
	if old == nil {
		return
	}
	tv := reflect.ValueOf(b).Elem()
	ov := reflect.ValueOf(old).Elem()
	structType := tv.Type()
	for i := 0; i < structType.NumField(); i++ {
		tf := tv.Field(i)
		bizType := tf.Type()
		if bizType.Kind() != reflect.Func {
			continue
		}
		of := ov.Field(i)
		if of.IsNil() {
			// 如果新方法为nil 则使用旧方法
			continue
		}
		if tf.IsNil() {
			// 旧方法如果为nil 则使用新方法
			tf.Set(of)
			continue
		}

		// Make a copy of tf for tf to call. (Otherwise it
		// creates a recursive call cycle and stack overflows)
		// 意思是新方法里面不能直接调用旧的方法
		tfCopy := reflect.ValueOf(tf.Interface())

		// We need to call both tf and of in some order. 合并之前的方法也会调用
		newFunc := reflect.MakeFunc(bizType, func(args []reflect.Value) []reflect.Value {
			tfCopy.Call(args)
			return of.Call(args)
		})
		tv.Field(i).Set(newFunc)
	}
}
