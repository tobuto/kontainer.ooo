package routing_test

import (
	"github.com/kontainerooo/kontainer.ooo/pkg/routing"
	"github.com/kontainerooo/kontainer.ooo/pkg/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routing", func() {
	Describe("Create Service", func() {
		It("Should create service", func() {
			routingService, err := routing.NewService(testutils.NewMockDB())
			Ω(err).ShouldNot(HaveOccurred())
			Expect(routingService).ToNot(BeZero())
		})

		It("Should return db error", func() {
			db := testutils.NewMockDB()
			db.SetError(1)
			_, err := routing.NewService(db)
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("Create Router Config", func() {
		db := testutils.NewMockDB()
		routingService, _ := routing.NewService(db)
		It("Should create RouterConfig with new RefID Name Pair", func() {
			err := routingService.CreateRouterConfig(&routing.RouterConfig{
				RefID: 0,
				Name:  "test",
			})
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("Should not create RouterConfig if RefID Name Pair already exists", func() {
			err := routingService.CreateRouterConfig(&routing.RouterConfig{
				RefID: 0,
				Name:  "test",
			})
			Ω(err).Should(HaveOccurred())
		})

		It("Should return error on db failure", func() {
			db.SetError(2)
			err := routingService.CreateRouterConfig(&routing.RouterConfig{})
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("Edit Router Config", func() {
		db := testutils.NewMockDB()
		routingService, _ := routing.NewService(db)
		It("Should change Router Config", func() {
			refID, name := uint(1), "test"
			routingService.CreateRouterConfig(&routing.RouterConfig{
				RefID: refID,
				Name:  name,
			})

			newName := "test2"
			err := routingService.EditRouterConfig(refID, name, &routing.RouterConfig{
				Name: newName,
			})
			Ω(err).ShouldNot(HaveOccurred())

			conf := &routing.RouterConfig{}
			routingService.GetRouterConfig(refID, newName, conf)
			Expect(conf.Name).To(BeEquivalentTo(newName))
		})

		It("Should prevent from changing the refID", func() {
			err := routingService.EditRouterConfig(1, "test2", &routing.RouterConfig{
				RefID: 2,
			})
			Ω(err).Should(HaveOccurred())
		})

		It("Should return error on db failure", func() {
			db.SetError(1)
			err := routingService.EditRouterConfig(1, "", &routing.RouterConfig{})
			Ω(err).Should(HaveOccurred())

			db.SetError(2)
			err = routingService.EditRouterConfig(1, "", &routing.RouterConfig{})
			Ω(err).Should(HaveOccurred())
		})
	})

	Describe("GetRouterConfig", func() {
		db := testutils.NewMockDB()
		routingService, _ := routing.NewService(db)
		It("Should fill RouterConfig struct", func() {
			refID, name := uint(1), "test"
			routingService.CreateRouterConfig(&routing.RouterConfig{
				RefID: refID,
				Name:  name,
			})

			conf := &routing.RouterConfig{}
			err := routingService.GetRouterConfig(refID, name, conf)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(conf.Name).To(BeEquivalentTo(name))
		})

		It("Should return error if ID does not exist", func() {
			err := routingService.GetRouterConfig(28, "", &routing.RouterConfig{})
			Ω(err).Should(BeEquivalentTo(testutils.ErrNotFound))
		})

		It("Should return error on db failure", func() {
			db.SetError(1)
			err := routingService.GetRouterConfig(1, "", &routing.RouterConfig{})
			Ω(err).Should(HaveOccurred())
		})
	})
})
