package admin

const styleMain = `
#rexray-banner-logo-div {
    height: 250px;
    display: inline-block;
}

#rexray-banner-logo {
    height: 100%;
}

#rexray-banner-logo-text {
    display: inline-block;
    vertical-align: top;
    line-height: 175px;
    height: 175px;
    padding-left: 20px;
    color: white;
    font-size: 4em;
    font-family: helvetica;
    text-shadow: 1px 1px 3px #fff;
}

#banner {
    position: fixed;
    top: 0;
    left: -5;
    right: -5;
    height: 175px;
    background-color: #000;
    box-shadow: 0px 5px 15px #000;
}

#main {
    width: 100%;
    margin-top: 215px;
}

#module-table {
    margin-left: auto;
    margin-right: auto;
    padding: 0;
    border: 0;
    border-collapse: collapse;
    border-spacing: 0;
}

#module-table th {
    padding-left: 15px;
    padding-right: 15px;
    text-align: left;
}

#module-table th.col-id {
    width: 35px;
}

#module-table th.col-name {
    width: 75px;
}

#module-table th.col-isup {
    width: 45px;
}

#module-table th.col-address {
}

#module-table th.col-description {
}

#module-table td {
    padding: 15px;
    text-align: left;
}

#module-table tbody tr:nth-child(even) {
    background-color: #a7d2f0
}

#module-table tbody tr:nth-child(odd) {
    background-color: #fff
}
#forkongithub a {
    background: #000;
    color: #fff;
    text-decoration: none;
    font-family: arial, sans-serif;
    text-align: center;
    font-weight: bold;
    padding: 5px 40px;
    font-size: 1rem;
    line-height: 2rem;
    position: relative;
    transition: 0.5s;
}

#forkongithub a:hover {
    background: #2c95dd;
    color: #fff;
}

#forkongithub a::before,
#forkongithub a::after {
    content: "";
    width: 100%;
    display: block;
    position: absolute;
    top: 1px;
    left: 0;
    height: 1px;
    background: #fff;
}

#forkongithub a::after {
    bottom: 1px;
    top: auto;
}

#forkongithub {
    position: fixed;
    display: block;
    top: 0;
    right: 0;
    width: 200px;
    overflow: hidden;
    height: 200px;
    z-index: 9999;
}

#forkongithub a {
    width: 200px;
    position: absolute;
    top: 60px;
    right: -60px;
    transform: rotate(45deg);
    -webkit-transform: rotate(45deg);
    -ms-transform: rotate(45deg);
    -moz-transform: rotate(45deg);
    -o-transform: rotate(45deg);
    box-shadow: 4px 4px 10px rgba(0, 0, 0, 0.8);
}

@media screen and (max-width:800px) {
    #forkongithub {
        display: none;
    }
}
`
