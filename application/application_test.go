package application

import (
	"summer/types"
	"testing"
	"github.com/stretchr/testify/assert"
)

type RouterConfig struct {
	types.Configuration
	*UserComponent `name:"*UserComponent"`
}

type HandlerConfig struct {
	types.Configuration
}

func (rc *RouterConfig) GetUserComponent() *UserComponent {
	return new(UserComponent)
}

func (hc *HandlerConfig) GetUserComponent() *UserComponent {
	return new(UserComponent)
}

type UserComponent struct {
	types.Component
	*RouterConfig
}

type App struct {
	*RouterConfig
	*HandlerConfig
}

func TestCreateApplication(t *testing.T) {
	app := CreateApplication(new(App))
	app.Run()
	for Type, valueMap := range app.BeanMap {
		t.Logf("Type: %s\n", Type.String())
		for name, value := range valueMap {
			t.Logf(" %s - Value canset: %s\n", name, value.Value.CanSet())
			t.Logf(" %s - Field canset: %s\n", name, value.Value.Field(0).CanSet())
			if Type.String() == "*application.Application" && Type.String() == name {
				assert.False(t, value.Value.Field(0).CanSet(), "Field of *application.Application should not CanSet")
			} else {
				assert.True(t, value.Value.Field(0).CanSet(), "Field of %s should CanSet", name)
			}

			assert.True(t, value.Value.CanSet(), "%s should CanSet", name)
		}
	}
}
