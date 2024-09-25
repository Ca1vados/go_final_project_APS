package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/siavoid/task-manager/entity"
	cnst "github.com/siavoid/task-manager/usecase/constants"
)

func (u *Usecase) CreateTask(task entity.Task) (int, error) {
	now := time.Now()

	// Проверяем обязательное поле Title
	if task.Title == "" {
		return 0, fmt.Errorf("Не указан заголовок задачи")
	}
	_, err := NextDate(now, now.Format(cnst.DateFormat), task.Repeat)
	if err != nil {
		return 0, err
	}

	// Проверяем и обрабатываем поле Date
	if task.Date == "" {
		task.Date = now.Format(cnst.DateFormat)
	} else {
		parsedDate, err := time.Parse(cnst.DateFormat, task.Date)
		if err != nil {
			return 0, fmt.Errorf("incorrect date format: %s", task.Date)
		}

		if parsedDate.Before(now) && parsedDate.Format(cnst.DateFormat) != now.Format(cnst.DateFormat) {
			if task.Repeat == "" {
				task.Date = now.Format(cnst.DateFormat)
			} else {
				nextDate, err := NextDate(now, task.Date, task.Repeat)
				if err != nil {
					return 0, err
				}
				task.Date = nextDate
			}
		}
	}
	taskId, err := u.db.CreateTask(task)
	return taskId, err
}

func (u *Usecase) GetAllTask() ([]entity.Task, error) {
	tasks, err := u.db.GetAllTasks()
	if err != nil {
		return make([]entity.Task, 0), err
	}

	if tasks == nil {
		return make([]entity.Task, 0), nil
	}

	n := 25
	if len(tasks) > n {
		tasks = tasks[:n]
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Date < tasks[j].Date
	})

	return tasks, err
}

func (u *Usecase) GetTask(id int) (entity.Task, error) {
	task, err := u.db.GetTask(id)

	return task, err
}

func (u *Usecase) UpdateTask(task entity.Task) error {
	// Проверяем обязательные поля
	if task.Title == "" {
		return errors.New("не указан заголовок задачи")
	}

	if task.ID == 0 {
		return errors.New("не указан id задачи")
	}

	// Проверяем формат даты
	if task.Date != "" {
		if _, err := time.Parse(cnst.DateFormat, task.Date); err != nil {
			return errors.New("некорректный формат даты")
		}
	}

	now := time.Now()

	_, err := NextDate(now, now.Format(cnst.DateFormat), task.Repeat)
	if err != nil {
		return err
	}

	_, err = time.Parse(cnst.DateFormat, task.Date)
	if err != nil {
		return err
	}

	err = u.db.UpdateTask(task)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("Задача не найдена")
		} else {
			fmt.Errorf("Ошибка обновления задачи")
		}
		return err
	}
	return err
}

func (u *Usecase) MarkTaskDone(id int) error {
	task, err := u.GetTask(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		// Удаляем одноразовую задачу
		if err := u.db.RemoveTask(id); err != nil {
			return err
		}
	} else {
		// Рассчитываем следующую дату выполнения для повторяющейся задачи
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return err
		}

		task.Date = nextDate
		if err := u.db.UpdateTask(task); err != nil {
			return err
		}
	}
	return nil
}

func (u *Usecase) DeleteTask(id int) error {
	err := u.db.RemoveTask(id)
	return err
}
