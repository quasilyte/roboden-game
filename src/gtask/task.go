package gtask

import (
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
)

type Task struct {
	ctx *TaskContext

	f func(ctx *TaskContext)

	completed    int32
	lastProgress float64

	EventProgress  gsignal.Event[TaskProgress]
	EventCompleted gsignal.Event[gsignal.Void]
}

type TaskProgress struct {
	Current float64
	Total   float64
}

type TaskContext struct {
	Progress TaskProgress

	elapsed float64
}

func (ctx *TaskContext) GetElapsedTime() float64 {
	return ctx.elapsed
}

func StartTask(f func(ctx *TaskContext)) *Task {
	task := &Task{
		ctx: &TaskContext{},
		f:   f,
	}
	return task
}

func (task *Task) Init(scene *ge.Scene) {
	go func() {
		task.f(task.ctx)
		task.EventCompleted.Emit(gsignal.Void{})
		atomic.StoreInt32(&task.completed, 1)
	}()
}

func (task *Task) Update(delta float64) {
	task.ctx.elapsed += delta

	progress := atomicLoadFloat64(&task.ctx.Progress.Current)
	if progress != task.lastProgress {
		task.lastProgress = progress
		task.EventProgress.Emit(task.ctx.Progress)
	}
}

func (task *Task) IsDisposed() bool {
	return task.completed != 0
}

func atomicLoadFloat64(x *float64) float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(x))))
}
