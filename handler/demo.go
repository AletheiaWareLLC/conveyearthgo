package handler

import (
	"aletheiaware.com/authgo"
	"aletheiaware.com/netgo"
	"aletheiaware.com/netgo/handler"
	"html/template"
	"log"
	"math"
	"net/http"
	"time"
)

func AttachDemoHandler(m *http.ServeMux, a authgo.Authenticator, ts *template.Template) {
	m.Handle("/demo", handler.Log(Demo(a, ts)))
}

func Demo(a authgo.Authenticator, ts *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		time1 := now.AddDate(0, 0, -7)
		time2 := now.AddDate(0, 0, -6)
		time3 := now.AddDate(0, 0, -5)
		time4 := now.AddDate(0, 0, -4)
		time5 := now.AddDate(0, 0, -3)
		time6 := now.AddDate(0, 0, -2)
		time7 := now.AddDate(0, 0, -1)
		time8 := now
		topic := "A Bedtime Story"
		content1 := "Once upon a time, in a land far away..."
		content2a := "...lived a huge green dragon whose name was Faye."
		content3a := "In a cave upon a horde of treasure she would lay..."
		content4a := "...enthralled by heaps of gold and glitter all day."
		content5a := "Through the dense fog of a bitter winter's night..."
		content6a := "...emerged a black stallion mounted by a silver knight."
		content7a := "With a shield in his left hand and a sword in his right..."
		content8a := "...the unwelcome arrival stood tall and looking for a fight."

		content2b := "...roamed an elephant, tall and gray."
		content3b := "As matriarch she lead, mile after mile..."
		content4b := "...until at a river, she met a crocodile."
		content5b := "With a mouthful of teeth and a tail like a whip..."
		content6b := "...the creature dared the herd to take a dip."
		content7b := "There was no way over and no way around..."
		content8b := "...the herd had to leave the safety of dry ground."
		toHTML := func(text string) template.HTML {
			return template.HTML("<p style=\"margin: 8px 0px;\">" + template.HTMLEscapeString(text) + "</p>")
		}
		toYield := func(responses ...int) (yield int) {
			for i, r := range responses {
				yield += r / int(math.Pow(2, float64(i+1)))
			}
			return
		}
		type GiftData struct {
			ConversationID int64
			MessageID      int64
			Author         *authgo.Account
			Amount         int64
			Created        time.Time
		}
		type MessageData struct {
			Created        time.Time
			ConversationID int64
			MessageID      int64
			Author         *authgo.Account
			Content        template.HTML
			Cost           int
			Yield          int
			Replies        []MessageData
			Gifts          []*GiftData
		}
		type ConversationData struct {
			MessageData
			Topic string
		}
		data := struct {
			Live                bool
			Step1, Step2, Step3 ConversationData
		}{
			Live: netgo.IsLive(),
			Step1: ConversationData{
				Topic: topic,
				MessageData: MessageData{
					Created: time1,
					Author:  &authgo.Account{Username: "Alice"},
					Content: toHTML(content1),
					Cost:    len([]byte(content1)),
				},
			},
			Step2: ConversationData{
				Topic: topic,
				MessageData: MessageData{
					Created: time1,
					Author:  &authgo.Account{Username: "Alice"},
					Content: toHTML(content1),
					Cost:    len([]byte(content1)),
					Yield:   toYield(len([]byte(content2a))) + toYield(len([]byte(content2b))),
					Replies: []MessageData{
						MessageData{
							Created: time2,
							Author:  &authgo.Account{Username: "Bob"},
							Content: toHTML(content2a),
							Cost:    len([]byte(content2a)),
						},
						MessageData{
							Created: time2,
							Author:  &authgo.Account{Username: "Beatrice"},
							Content: toHTML(content2b),
							Cost:    len([]byte(content2b)),
						},
					},
				},
			},
			Step3: ConversationData{
				Topic: topic,
				MessageData: MessageData{
					Created: time1,
					Author:  &authgo.Account{Username: "Alice"},
					Content: toHTML(content1),
					Cost:    len([]byte(content1)),
					Yield:   toYield(len([]byte(content2a)), len([]byte(content3a)), len([]byte(content4a)), len([]byte(content5a)), len([]byte(content6a)), len([]byte(content7a)), len([]byte(content8a))) + toYield(len([]byte(content2b)), len([]byte(content3b)), len([]byte(content4b)), len([]byte(content5b)), len([]byte(content6b)), len([]byte(content7b)), len([]byte(content8b))),
					Replies: []MessageData{
						MessageData{
							Created: time2,
							Author:  &authgo.Account{Username: "Bob"},
							Content: toHTML(content2a),
							Cost:    len([]byte(content2a)),
							Yield:   toYield(len([]byte(content3a)), len([]byte(content4a)), len([]byte(content5a)), len([]byte(content6a)), len([]byte(content7a)), len([]byte(content8a))),
							Replies: []MessageData{
								MessageData{
									Created: time3,
									Author:  &authgo.Account{Username: "Claire"},
									Content: toHTML(content3a),
									Cost:    len([]byte(content3a)),
									Yield:   toYield(len([]byte(content4a)), len([]byte(content5a)), len([]byte(content6a)), len([]byte(content7a)), len([]byte(content8a))),
									Replies: []MessageData{
										MessageData{
											Created: time4,
											Author:  &authgo.Account{Username: "Daniel"},
											Content: toHTML(content4a),
											Cost:    len([]byte(content4a)),
											Yield:   toYield(len([]byte(content5a)), len([]byte(content6a)), len([]byte(content7a)), len([]byte(content8a))),
											Replies: []MessageData{
												MessageData{
													Created: time5,
													Author:  &authgo.Account{Username: "Elizabeth"},
													Content: toHTML(content5a),
													Cost:    len([]byte(content5a)),
													Yield:   toYield(len([]byte(content6a)), len([]byte(content7a)), len([]byte(content8a))),
													Replies: []MessageData{
														MessageData{
															Created: time6,
															Author:  &authgo.Account{Username: "Fred"},
															Content: toHTML(content6a),
															Cost:    len([]byte(content6a)),
															Yield:   toYield(len([]byte(content7a)), len([]byte(content8a))),
															Replies: []MessageData{
																MessageData{
																	Created: time7,
																	Author:  &authgo.Account{Username: "Ginny"},
																	Content: toHTML(content7a),
																	Cost:    len([]byte(content7a)),
																	Yield:   toYield(len([]byte(content8a))),
																	Replies: []MessageData{
																		MessageData{
																			Created: time8,
																			Author:  &authgo.Account{Username: "Helen"},
																			Content: toHTML(content8a),
																			Cost:    len([]byte(content8a)),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						MessageData{
							Created: time2,
							Author:  &authgo.Account{Username: "Beatrice"},
							Content: toHTML(content2b),
							Cost:    len([]byte(content2b)),
							Yield:   toYield(len([]byte(content3b)), len([]byte(content4b)), len([]byte(content5b)), len([]byte(content6b)), len([]byte(content7b)), len([]byte(content8b))),
							Replies: []MessageData{
								MessageData{
									Created: time3,
									Author:  &authgo.Account{Username: "Charlie"},
									Content: toHTML(content3b),
									Cost:    len([]byte(content3b)),
									Yield:   toYield(len([]byte(content4b)), len([]byte(content5b)), len([]byte(content6b)), len([]byte(content7b)), len([]byte(content8b))),
									Replies: []MessageData{
										MessageData{
											Created: time4,
											Author:  &authgo.Account{Username: "Dina"},
											Content: toHTML(content4b),
											Cost:    len([]byte(content4b)),
											Yield:   toYield(len([]byte(content5b)), len([]byte(content6b)), len([]byte(content7b)), len([]byte(content8b))),
											Replies: []MessageData{
												MessageData{
													Created: time5,
													Author:  &authgo.Account{Username: "Edgar"},
													Content: toHTML(content5b),
													Cost:    len([]byte(content5b)),
													Yield:   toYield(len([]byte(content6b)), len([]byte(content7b)), len([]byte(content8b))),
													Replies: []MessageData{
														MessageData{
															Created: time6,
															Author:  &authgo.Account{Username: "Fiona"},
															Content: toHTML(content6b),
															Cost:    len([]byte(content6b)),
															Yield:   toYield(len([]byte(content7b)), len([]byte(content8b))),
															Replies: []MessageData{
																MessageData{
																	Created: time7,
																	Author:  &authgo.Account{Username: "George"},
																	Content: toHTML(content7b),
																	Cost:    len([]byte(content7b)),
																	Yield:   toYield(len([]byte(content8b))),
																	Replies: []MessageData{
																		MessageData{
																			Created: time8,
																			Author:  &authgo.Account{Username: "Harold"},
																			Content: toHTML(content8b),
																			Cost:    len([]byte(content8b)),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		if err := ts.ExecuteTemplate(w, "demo.go.html", data); err != nil {
			log.Println(err)
			return
		}
	})
}
