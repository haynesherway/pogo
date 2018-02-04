package pogo

import (
    "fmt"
    "github.com/bwmarrin/discordgo"
    "strconv"
    "strings"
    "sync"
    "time"
)

const (
    status_startiing int = iota
    status_expecting_pokemon 
    status_got_pokemon
    status_expecting_cp
    status_got_cp
    status_expecting_level
    status_calculating
    status_done
)

type IVCalculator struct {
    Session *discordgo.Session
    InputChannel chan interface{}
    RunningCalculations map[string]*IVCalculation
    lock *sync.RWMutex
}

type IVCalculation struct {
    Session *discordgo.Session
    User *discordgo.User
    ChannelID string
    Pokemon *Pokemon
    IV  *ivStat
    Channel chan interface{}
    Status int
}

func StartIVCalculator(s *discordgo.Session) (*IVCalculator) {
    ivCalculator := IVCalculator{
        Session: s,
        InputChannel: make(chan interface{}),
        RunningCalculations: make(map[string]*IVCalculation),
        lock: &sync.RWMutex{},
    }

    go func() {
        for incoming := range ivCalculator.InputChannel {
            if m, ok := incoming.(*discordgo.MessageCreate); ok {
                if !ivCalculator.IsRunning(m.Author.ID) {
                    //Start New Calculation
                    ivCalculator.Start(m)
                } else {
                    ivCalc := ivCalculator.GetCalculation(m.Author.ID)
                    if ivCalc.Status == status_done {
                        ivCalculator.Stop(m.Author.ID)
                        ivCalculator.Start(m)
                    } else if ivCalc.ChannelID == m.ChannelID {
                        ivCalc.Channel <- m.Content
                    }
                }
            }
        }
    }()
    return &ivCalculator
}

func (calc *IVCalculator) Start(m *discordgo.MessageCreate) {
    thisCalculation := &IVCalculation{
        Session: calc.Session,
        User: m.Author,
        ChannelID: m.ChannelID,
        Channel: make(chan interface{}),
    }
    calc.lock.Lock()
    calc.RunningCalculations[m.Author.ID] = thisCalculation
    calc.lock.Unlock()
    
    thisCalculation.AskQuestion()
    
    go func() {
        timeout := time.After(1 * time.Minute)
        for {
            select {
                case msg := <- thisCalculation.Channel:
                    if message, ok := msg.(string); ok {
                        thisCalculation.GetResponse(message)
                    }
                    
                case <-timeout:
                    if thisCalculation.Status == status_done {
                        //YAY it completed sucessfully
                    } else {
                        thisCalculation.PrintToDiscord("Unable to process your IV Calculation, please try again.")
                    }
                    calc.Stop(m.Author.ID)
                    return
            }
        }
    }()
    
    return
}

func (calc *IVCalculator) Stop(userID string) {
    calc.lock.Lock()
    delete(calc.RunningCalculations, userID)
    calc.lock.Unlock()
}

func (calc *IVCalculator) IsRunning(userID string) bool {
    calc.lock.RLock()
    defer calc.lock.RUnlock()
    if _, ok := calc.RunningCalculations[userID]; ok {
        return true
    }
    return false
}

func (calc *IVCalculator) GetCalculation(userID string) (*IVCalculation) {
    calc.lock.RLock()
    defer calc.lock.RUnlock()
    if ivCalc, ok := calc.RunningCalculations[userID]; ok {
        return ivCalc
    }
    return nil
}

func (ivCalc *IVCalculation) AskQuestion() {
    
    switch ivCalc.Status {
        case status_startiing:
            ivCalc.PrintToDiscord("Enter pokemon name.")
        case status_got_pokemon:
            ivCalc.PrintToDiscord("Enter CP.")
        case status_got_cp
:
            ivCalc.PrintToDiscord("Enter level.")
            
    }
    
    ivCalc.Status++
    fmt.Println(ivCalc.Status)
    return
}

func (ivCalc *IVCalculation) GetResponse(m string) {
    
    switch ivCalc.Status {
    case status_expecting_pokemon:
        if p, ok := PokemonMap[strings.ToLower(m)]; ok {
            ivCalc.Pokemon = &p
            ivCalc.Status++
            ivCalc.AskQuestion()
        } else {
            ivCalc.PrintToDiscord("Unrecognized pokemon. Try again.")
        }
    case status_expecting_cp:
        if cp, err := strconv.Atoi(m); err == nil {
            ivCalc.IV.CP = cp
            ivCalc.Status++
            ivCalc.AskQuestion()
        } else {
            ivCalc.PrintToDiscord(fmt.Sprintf("CP must be an integer, got %s. Try again.", m))
        }
    case status_expecting_level:
        if lvl, err := strconv.ParseFloat(m, 64); err == nil {
            ivCalc.IV.Level = lvl
            ivCalc.Status++
        } else {
            ivCalc.PrintToDiscord(fmt.Sprintf("Level must be an integer, got %s. Try again.", m))
        }
    }
    
    if ivCalc.Status == status_calculating {
        ivCalc.Calculate()
    }
    
    return
}

func (ivCalc *IVCalculation) Calculate() {
    msg := ivCalc.Pokemon.getIV(ivCalc.IV)
    ivCalc.PrintToDiscord(msg)
    ivCalc.Status++
}

func (ivCalc *IVCalculation) PrintToDiscord(m string) {
    _, _ = ivCalc.Session.ChannelMessageSend(ivCalc.ChannelID, ivCalc.User.Mention() + " " + m)
    return
}