package unit

import (
	"github.com/nikita-itmo-gh-acc/car_estimator_authorization/tests/unit/mocks"
)

type MockDependencies struct {
	userProvider   	*mocks.IUserProvider
	userSaver		*mocks.IUserSaver
	sessionProvider *mocks.ISessionProvider
	sessionSaver 	*mocks.ISessionSaver
	sessionRemover 	*mocks.ISessionRemover
}

type TestCase struct {
	name       string
	args       any
	setupMocks func(*MockDependencies)
	wantErr    bool
}
